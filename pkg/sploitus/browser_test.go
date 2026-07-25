package sploitus

import "testing"

func TestParseExploitIDFromHref(t *testing.T) {
	tests := []struct {
		href     string
		expected string
	}{
		{"/exploit?id=0147E6AA-6963-51CE-90F9-420346FA917B", "0147E6AA-6963-51CE-90F9-420346FA917B"},
		{"/exploit?id=CVE-2023-12345", "CVE-2023-12345"},
		{"/exploit?id=SAINT:2CEDD0194C77120545A6315E534CFE66", "SAINT:2CEDD0194C77120545A6315E534CFE66"},
		{"/exploit?id=", ""},
		{"/exploit", ""},
		{"", ""},
		{"no-id-here", ""},
	}

	for _, tt := range tests {
		result := parseExploitIDFromHref(tt.href)
		if result != tt.expected {
			t.Errorf("parseExploitIDFromHref(%q) = %q, want %q", tt.href, result, tt.expected)
		}
	}
}