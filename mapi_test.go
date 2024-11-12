package tnef

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMAPIAttribute_Value(t *testing.T) {
	m := &MAPIAttribute{}
	assert.False(t, m.HasName("not set"))

	m.Names = []string{
		"x-test-header",
		"second-name",
	}

	assert.True(t, m.HasName("X-Test-Header"))
	assert.False(t, m.HasName("X-Test-Header-1"))
	assert.True(t, m.HasName("second-name"))
}

func TestMAPIAttribute_AsString(t *testing.T) {
	m := &MAPIAttribute{}
	_, err := m.AsString()
	assert.Error(t, err)

	m.Type = szmapiUnicodeString
	m.Data = []byte{
		97, 0, 112, 0, 112, 0, 108, 0, 105, 0, 99, 0, 97, 0, 116, 0,
		105, 0, 111, 0, 110, 0, 47, 0, 112, 0, 107, 0, 99, 0, 115, 0,
		55, 0, 45, 0, 109, 0, 105, 0, 109, 0, 101, 0, 59, 0, 115, 0,
		109, 0, 105, 0, 109, 0, 101, 0, 45, 0, 116, 0, 121, 0, 112,
		0, 101, 0, 61, 0, 101, 0, 110, 0, 118, 0, 101, 0, 108, 0,
		111, 0, 112, 0, 101, 0, 100, 0, 45, 0, 100, 0, 97, 0, 116, 0,
		97, 0, 59, 0, 110, 0, 97, 0, 109, 0, 101, 0, 61, 0, 115, 0,
		109, 0, 105, 0, 109, 0, 101, 0, 46, 0, 112, 0, 55, 0, 109,
		0, 0, 0,
	}
	value, err := m.AsString()
	assert.NoError(t, err)
	assert.Equal(t, "application/pkcs7-mime;smime-type=enveloped-data;name=smime.p7m", value)

	m.Type = 0x48
	m.Data = []byte{91, 0, 49, 0, 48, 0, 46, 0, 49, 0, 50, 0, 46, 0, 48, 0, 46, 0, 51, 0, 57, 0, 93, 0, 0, 0}
	_, err = m.AsString()
	assert.Error(t, err)
}
