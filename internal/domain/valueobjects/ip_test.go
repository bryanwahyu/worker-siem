package valueobjects_test

import (
	"testing"

	"github.com/bryanwahyu/worker-siem/internal/domain/valueobjects"
)

func TestNewIP(t *testing.T) {
	ip, err := valueobjects.NewIP("192.168.1.1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ip.String() != "192.168.1.1" {
		t.Fatalf("unexpected string: %s", ip.String())
	}
}

func TestNewIPInvalid(t *testing.T) {
	if _, err := valueobjects.NewIP("invalid"); err == nil {
		t.Fatalf("expected error for invalid IP")
	}
}
