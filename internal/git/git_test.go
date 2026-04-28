package git

import (
	"testing"
)

func TestGetCurrentSHA(t *testing.T) {
	sha, err := GetCurrentSHA(".")
	if err != nil {
		t.Fatalf("GetCurrentSHA failed: %v", err)
	}

	if len(sha) == 0 {
		t.Error("Returned SHA is empty")
	}
}
