package gitops

import (
	"testing"
)

func TestValidateRepoURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid https", "https://github.com/user/repo.git", false},
		{"valid ssh", "git@github.com:user/repo.git", false},
		{"invalid semicolon", "https://github.com/user/repo.git;rm -rf /", true},
		{"invalid pipe", "https://github.com/user/repo.git|cat /etc/passwd", true},
		{"invalid ampersand", "https://github.com/user/repo.git&whoami", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRepoURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRepoURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepoKey(t *testing.T) {
	key1 := repoKey("https://github.com/user/repo.git")
	key2 := repoKey("https://github.com/user/repo.git")
	if key1 != key2 {
		t.Error("same URL should generate same key")
	}
	if len(key1) != 40 {
		t.Errorf("expected SHA1 hash length 40, got %d", len(key1))
	}
}

func TestCompactStrings(t *testing.T) {
	input := []string{"", "a", "  ", "b", "a"}
	result := compactStrings(input)
	if len(result) != 2 {
		t.Errorf("expected 2 unique non-empty strings, got %d", len(result))
	}
}

func TestUniqueStrings(t *testing.T) {
	input := []string{"a", "b", "a", "c"}
	result := uniqueStrings(input)
	if len(result) != 3 {
		t.Errorf("expected 3 unique strings, got %d", len(result))
	}
}
