package main

import (
	"context"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/desktop/internal/update"
)

// updater_app.go exposes update detection while keeping in-app installation off.

// Version returns the build version injected via -ldflags (see main.go).
func (a *App) Version() string { return version }

func (a *App) CheckUpdate() (*UpdateInfo, error) {
	client, err := httpClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(a.reqCtx(), 10*time.Second)
	defer cancel()
	release, err := fetchLatestDesktopRelease(ctx, client)
	if err != nil {
		return &UpdateInfo{Current: version, CanSelfUpdate: false, DownloadURL: ghReleasesBase, Err: err.Error()}, nil
	}
	info := evaluateGitHubRelease(version, release)
	if info.DownloadURL == "" {
		info.DownloadURL = ghReleasesBase
	}
	a.updateMu.Lock()
	a.updateURL = info.DownloadURL
	a.updateMu.Unlock()
	return &info, nil
}

func (a *App) OpenDownloadPage() {
	if a.ctx != nil {
		a.updateMu.RLock()
		url := a.updateURL
		a.updateMu.RUnlock()
		if url == "" {
			url = ghReleasesBase
		}
		wruntime.BrowserOpenURL(a.ctx, url)
	}
}

// ApplyUpdate is disabled in O.R.C.A builds.
func (a *App) ApplyUpdate() error {
	return nil
}

// downloadVerify downloads the asset (streaming progress), verifies its minisign
// signature against the embedded public key, then its sha256. It returns the
// verified bytes and never touches disk on a bad signature.
func (a *App) downloadVerify(asset update.Asset) ([]byte, error) {
	c, err := httpClient()
	if err != nil {
		return nil, err
	}
	data, err := download(a.reqCtx(), c, asset.URL, asset.Size, func(rcv, total int64) {
		a.emitProgress("downloading", rcv, total, "")
	})
	if err != nil {
		return nil, err
	}
	a.emitProgress("verifying", asset.Size, asset.Size, "")
	sig, err := fetchBytes(a.reqCtx(), c, asset.Sig)
	if err != nil {
		return nil, err
	}
	if err := update.Verify(data, sig); err != nil {
		return nil, err
	}
	if err := checkSHA256(data, asset.SHA256); err != nil {
		return nil, err
	}
	return data, nil
}

// reqCtx is the context for updater HTTP calls — the Wails context once startup has
// run, else Background (CheckUpdate may, in theory, be reached before startup).
func (a *App) reqCtx() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}

func (a *App) emitProgress(phase string, received, total int64, errMsg string) {
	if a.ctx == nil {
		return
	}
	wruntime.EventsEmit(a.ctx, "updater:progress", updateProgress{
		Phase: phase, Received: received, Total: total, Err: errMsg,
	})
}

// failUpdate emits an error progress event and returns the error to the caller.
func (a *App) failUpdate(err error) error {
	a.emitProgress("error", 0, 0, err.Error())
	return err
}
