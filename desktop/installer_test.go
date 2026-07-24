package main

import (
	"os"
	"strings"
	"testing"
)

func TestWindowsInstallerOffersShortcutAndLaunchChoices(t *testing.T) {
	body, err := os.ReadFile("build/windows/installer/project.nsi")
	if err != nil {
		t.Fatal(err)
	}
	script := string(body)
	for _, want := range []string{
		"Page custom InstallOptionsPage InstallOptionsPageLeave",
		"MUI_FINISHPAGE_RUN",
		"CreateDesktopShortcut == ${BST_CHECKED}",
		"Delete \"$DESKTOP\\${INFO_PRODUCTNAME}.lnk\"",
		"IfFileExists \"$DESKTOP\\${INFO_PRODUCTNAME}.lnk\"",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("installer is missing %q", want)
		}
	}
}
