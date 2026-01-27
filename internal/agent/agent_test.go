package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocalIP(t *testing.T) {
	ip := getLocalIP()

	assert.NotEmpty(t, ip, "getLocalIP should return a non-empty IP")
	assert.NotEqual(t, "127.0.0.1", ip, "getLocalIP should return non-loopback IP when available")
}
