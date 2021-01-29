package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterfaceToString(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input  interface{}
		output string
		isErr  bool
	}{
		"empty":   {input: nil, isErr: true},
		"struct":  {input: map[string]string{}, isErr: true},
		"string":  {input: "string", output: "string"},
		"int":     {input: 10, output: "10"},
		"int64":   {input: int64(10), output: "10"},
		"float32": {input: float32(10), output: "10"},
		"float64": {input: float64(10), output: "10"},
		"bool":    {input: true, output: "true"},
	}

	for _, t := range tests {
		v, err := InterfaceToString(t.input)
		assert.Equal(t.isErr, err != nil)
		assert.Equal(t.output, v)
	}
}

func TestInterfaceTotMap(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		input  interface{}
		output map[string]string
		isErr  bool
	}{
		"err-1": {
			input: struct {
				a string
			}{}, isErr: true,
		},
		"err-2": {
			input: map[string]interface{}{"aa": "bb", "cc": struct{ aa string }{}}, isErr: true,
		},
		"map[string]string": {
			input: map[string]string{"aa": "bb", "cc": "dd"}, output: map[string]string{"aa": "bb", "cc": "dd"},
		},
		"map[string]int": {
			input: map[string]int{"aa": 11, "cc": 22}, output: map[string]string{"aa": "11", "cc": "22"},
		},
		"map[int]string": {
			input: map[int]string{ 11:"aa", 22: "cc"}, output: map[string]string{"11": "aa", "22": "cc"},
		},
		"map[interface]interface": {
			input: map[interface{}]interface{}{"aa": 11, "cc": 22}, output: map[string]string{"aa": "11", "cc": "22"},
		},
	}

	for _, t := range tests {
		output, err := InterfaceTotMap(t.input)
		assert.Equal(t.isErr, err != nil)
		if err == nil {
			for k, v := range output {
				assert.Equal(t.output[k], v)
			}
		}
	}
}
