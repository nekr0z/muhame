package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type cfg struct {
	String string
}

func TestMissing(t *testing.T) {
	c := cfg{}
	filename := "config.json"
	err := parseConfig(c, filename)
	assert.Error(t, err)
}

func TestFailing(t *testing.T) {
	c := cfg{}
	filename := "config_test.go"
	err := parseConfig(c, filename)
	assert.Error(t, err)
}
