package valueobjects

import (
	"fmt"
	"net"
)

// IP wraps a parsed IP address enforcing validation at creation time.
type IP struct {
	value net.IP
}

// NewIP validates the textual representation and returns an IP value object.
func NewIP(raw string) (IP, error) {
	ip := net.ParseIP(raw)
	if ip == nil {
		return IP{}, fmt.Errorf("invalid ip address: %s", raw)
	}
	return IP{value: ip}, nil
}

// String returns the canonical textual representation.
func (i IP) String() string {
	return i.value.String()
}

// IsZero indicates whether the IP has been initialized.
func (i IP) IsZero() bool {
	return len(i.value) == 0
}

// ToNet exposes the underlying net.IP for infrastructure use.
func (i IP) ToNet() net.IP {
	cp := make(net.IP, len(i.value))
	copy(cp, i.value)
	return cp
}
