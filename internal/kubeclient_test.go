package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewKubeClient(t *testing.T) {
	assert := assert.New(t)

	cfgPath, err := KubeConfigPath()
	if err != nil {
		return
	}

	kubeConfig, err := NewKubeConfig(cfgPath)
	assert.NoError(err)

	tests := map[string]struct {
		input *KubeConfig
		isErr bool
	}{
		"error-1":   {isErr: true},
		"success-1": {input: kubeConfig},
	}

	for _, t := range tests {
		_, err := NewKubeClient(t.input)
		assert.Equal(t.isErr, err != nil)
	}
}
