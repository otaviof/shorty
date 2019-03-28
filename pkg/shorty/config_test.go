package shorty

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var config *Config

func TestNewConfig(t *testing.T) {
	config = NewConfig()
}

func TestConfigValidate(t *testing.T) {
	err := config.Validate()
	assert.Nil(t, err)

	config.Address = ""
	err = config.Validate()
	assert.NotNil(t, err)
}
