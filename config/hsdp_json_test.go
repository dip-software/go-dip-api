package config_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dip-software/go-dip-api/config"

	"github.com/stretchr/testify/assert"
)

func TestHSDPJSONContent(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !assert.True(t, ok) {
		return
	}
	basePath := filepath.Dir(filename)
	hsdpTomlFile := filepath.Join(basePath, "hsdp.json")
	data, err := os.ReadFile(hsdpTomlFile)
	if !assert.Nil(t, err) {
		return
	}
	configReader := bytes.NewReader(data)

	c, err := config.New(config.FromReader(configReader))
	if !assert.Nil(t, err) {
		return
	}
	assert.Less(t, 0, len(c.Region("us-east").Services()))
	assert.Less(t, 0, len(c.Region("eu-west").Env("client-test").Service("pki").URL))
}
