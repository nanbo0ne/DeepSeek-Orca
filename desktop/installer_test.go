package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestWindowsInstallerOffersShortcutAndLaunchChoices(t *testing.T) {
	body, err := os.ReadFile("build/windows/installer/project.nsi")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(body, []byte{0xEF, 0xBB, 0xBF}) {
		t.Fatal("installer script must use a UTF-8 BOM so makensis decodes custom Chinese text correctly")
	}
	script := string(body)
	for _, want := range []string{
		"Page custom InstallOptionsPage InstallOptionsPageLeave",
		"MUI_FINISHPAGE_RUN",
		"运行 ${INFO_PRODUCTNAME}",
		"选择安装选项",
		"创建桌面快捷方式",
		"CreateDesktopShortcut == ${BST_CHECKED}",
		"Delete \"$DESKTOP\\${INFO_PRODUCTNAME}.lnk\"",
		"IfFileExists \"$DESKTOP\\${INFO_PRODUCTNAME}.lnk\"",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("installer is missing %q", want)
		}
	}
}
