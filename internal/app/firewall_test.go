package app

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/thomasilliet/gfw/internal/store"
)

func TestNormalizeFormat(t *testing.T) {
	for _, test := range []struct {
		input, want string
		ok          bool
	}{{"", "json", true}, {"JSON", "json", true}, {"txt", "txt", true}, {"yaml", "yaml", false}} {
		got, ok := normalizeFormat(test.input)
		if got != test.want || ok != test.ok {
			t.Fatalf("normalizeFormat(%q)=(%q,%v), want (%q,%v)", test.input, got, ok, test.want, test.ok)
		}
	}
}

func TestEncodeListFormatsAndStableETag(t *testing.T) {
	items := []store.EffectiveIP{
		{IPv4: "10.0.0.2", SourceType: "machine", SourceID: "m1"},
		{IPv4: "10.0.0.1", SourceType: "trusted", SourceID: "t1"},
		{IPv4: "10.0.0.2", SourceType: "trusted", SourceID: "t2"},
	}
	text, textETag, err := encodeList("payments", "txt", items)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(text), "10.0.0.1\n10.0.0.2\n"; got != want {
		t.Fatalf("txt=%q want %q", got, want)
	}
	jsonOne, jsonETagOne, err := encodeList("payments", "json", items)
	if err != nil {
		t.Fatal(err)
	}
	jsonTwo, jsonETagTwo, err := encodeList("payments", "json", items)
	if err != nil {
		t.Fatal(err)
	}
	if jsonETagOne != jsonETagTwo {
		t.Fatalf("unstable ETag: %s != %s", jsonETagOne, jsonETagTwo)
	}
	if textETag == jsonETagOne {
		t.Fatal("format-specific ETags must differ")
	}
	var response firewallResponse
	if err := json.Unmarshal(jsonOne, &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Items) != 3 || !strings.Contains(string(jsonTwo), `"SourceType"`) {
		t.Fatal("JSON must preserve separate provenance entries")
	}
}
