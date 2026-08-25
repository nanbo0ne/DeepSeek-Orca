package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeLocalAICatalogViewUsesJSONArrays(t *testing.T) {
	view := normalizeLocalAICatalogView(LocalAICatalogView{})
	data, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"models", "runtimes", "installedModels", "downloads", "gpus"} {
		if strings.Contains(string(data), `"`+field+`":null`) {
			t.Fatalf("%s must serialize as an array: %s", field, data)
		}
	}
}
