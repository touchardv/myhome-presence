package tplink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMD5(t *testing.T) {
	r := md5Func("foobar")
	assert.Equal(t, "3858f62230ac3c915f300c664312c63f", r)
}

func TestRE450Token(t *testing.T) {
	r := re450Token("foobar", "nonce")
	assert.Equal(t, "0AA33C47954A9397B25860B8C0DAF623", r)
}
