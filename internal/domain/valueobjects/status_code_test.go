package valueobjects_test

import (
	"testing"

	"github.com/bryanwahyu/worker-siem/internal/domain/valueobjects"
)

func TestNewStatusCode(t *testing.T) {
	sc, err := valueobjects.NewStatusCode(200)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sc.Int() != 200 {
		t.Fatalf("expected 200, got %d", sc.Int())
	}
}

func TestNewStatusCodeInvalid(t *testing.T) {
	if _, err := valueobjects.NewStatusCode(10); err == nil {
		t.Fatalf("expected error for invalid status code")
	}
}
