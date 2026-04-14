package app

import "testing"

func TestParseRunOptions_ReadsCustomConfigPath(t *testing.T) {
	t.Parallel()

	options, err := parseRunOptions([]string{"--config", "/tmp/custom.yaml"})
	if err != nil {
		t.Fatalf("parseRunOptions returned error: %v", err)
	}

	if options.ConfigPath != "/tmp/custom.yaml" {
		t.Fatalf("expected config path to be parsed, got %q", options.ConfigPath)
	}
}

func TestParseRunOptions_UsesEmptyConfigPathByDefault(t *testing.T) {
	t.Parallel()

	options, err := parseRunOptions(nil)
	if err != nil {
		t.Fatalf("parseRunOptions returned error: %v", err)
	}

	if options.ConfigPath != "" {
		t.Fatalf("expected empty config path by default, got %q", options.ConfigPath)
	}
}
