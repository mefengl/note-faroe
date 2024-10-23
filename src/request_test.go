package main

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyRequestSecret(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "abc")
	assert.Equal(t, true, verifyRequestSecret([]byte{}, r))

	r = httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "")
	assert.Equal(t, true, verifyRequestSecret([]byte{}, r))

	r = httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "abc")
	assert.Equal(t, true, verifyRequestSecret([]byte("abc"), r))

	r = httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "")
	assert.Equal(t, false, verifyRequestSecret([]byte("abc"), r))

	r = httptest.NewRequest("GET", "/", nil)
	assert.Equal(t, false, verifyRequestSecret([]byte("abc"), r))
}
