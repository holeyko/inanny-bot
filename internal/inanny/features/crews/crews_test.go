package crew

import (
	"strings"
	"testing"
)

func TestValidateNameSupportsUnicode(t *testing.T) {
	validNames := []string{
		"команда",
		"チーム_123",
		"नमस्ते",
		strings.Repeat("a", 128),
	}

	for _, name := range validNames {
		if err := ValidateName(name); err != nil {
			t.Errorf("ValidateName(%q) error = %v", name, err)
		}
	}
}

func TestValidateNameRejectsInvalidLengthAndCharacters(t *testing.T) {
	tests := []string{
		strings.Repeat("a", 129),
		"name with spaces",
		"name-with-dashes",
	}

	for _, name := range tests {
		if err := ValidateName(name); err == nil {
			t.Errorf("ValidateName(%q) error = nil, want validation error", name)
		}
	}
}
