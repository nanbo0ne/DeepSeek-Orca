package main

import (
	"context"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"deepseek-orca/desktop/internal/update"
)

// updater_app.go keeps the former updater command surface for compatibility.
// DeepSeek-Orca disables update checks and update application in the product UI.

// Version returns the build version injected via -ldflags (see main.go).
func (a *App) Version() string { return version }

// CheckUpdate is disabled in DeepSeek-Orca builds; the product no longer performs
// automatic or manual update checks.
func (a *App) CheckUpdate() (*UpdateInfo, error) {
	return nil, nil
}

// OpenDownloadPage is retained for API compatibility.
func (a *App) OpenDownloadPage() {
	if a.ctx != nil {
		wruntime.BrowserOpenURL(a.ctx, "https://github.com/nanbo0ne/DeepSeek-Orca")
	}
}

// ApplyUpdate is disabled in DeepSeek-Orca builds.
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
