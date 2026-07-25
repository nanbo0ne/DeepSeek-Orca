package main

import (
	"os"
	"strings"
	"testing"
)

func TestWindowsReleaseRequiresTwoStageSignPathSigning(t *testing.T) {
	body, err := os.ReadFile("../.github/workflows/release-desktop.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(body)
	for _, want := range []string{
		"Require SignPath release signing",
		"unsigned Windows releases are forbidden",
		"Upload unsigned Windows app for SignPath",
		"Repackage Windows installer with signed app",
		"Upload unsigned Windows installer for SignPath",
		"Verify signed Windows release artifacts",
		"TimeStamperCertificate",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("release workflow is missing %q", want)
		}
	}
	if got := strings.Count(workflow, "uses: signpath/github-action-submit-signing-request@v2"); got != 2 {
		t.Fatalf("SignPath request count = %d, want 2 (application and installer)", got)
	}
	if got := strings.Count(workflow, "artifact-configuration-slug: windows-executable"); got != 2 {
		t.Fatalf("SignPath artifact configuration count = %d, want 2", got)
	}
}

func TestSignPathArtifactConfigurationSignsPEFiles(t *testing.T) {
	body, err := os.ReadFile("../.signpath/artifact-configurations/windows-executable.xml")
	if err != nil {
		t.Fatal(err)
	}
	configuration := string(body)
	for _, want := range []string{"<zip-file>", `<pe-file path="*.exe">`, "<authenticode-sign />"} {
		if !strings.Contains(configuration, want) {
			t.Fatalf("SignPath artifact configuration is missing %q", want)
		}
	}
}
