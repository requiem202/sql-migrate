package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	ConfigFile = "../test-integration/dbconfig.yml"
	cfg, err := ReadConfig()

	assert.Nil(t, err)
	assert.Len(t, cfg, 5)
}
func TestReadConfigToml(t *testing.T) {
	ConfigFile = "../test-integration/dbconfig.toml"
	cfg, err := ReadConfigToml()

	assert.Nil(t, err)
	assert.Len(t, cfg, 5)
}
