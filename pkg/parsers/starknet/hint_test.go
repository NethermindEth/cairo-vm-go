package vm

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestParsingCoreHint(t *testing.T) {
	v := validator.New()

	testData := []byte(`{
		"TestLessThanOrEqual": {
			"lhs": { "Immediate": "0x95ec" },
			"rhs": { "Deref": { "register": "FP", "offset": -6 } },
			"dst": { "register": "AP", "offset": 1 }
		} 
	}`)

	var hint Hint
	err := json.Unmarshal(testData, &hint)
	assert.NoError(t, err)

	h, ok := hint.Args.(*TestLessThanOrEqual)
	assert.True(t, ok)

	assert.NoError(t, v.Struct(h))
}

func TestParseHintWithDoubleDeref(t *testing.T) {
	v := validator.New()

	testData := []byte(`{
		"TestLessThanOrEqual": {
			"lhs": { "DoubleDeref": [ { "register": "AP", "offset": 1 }, 1 ] },
			"rhs": { "Deref": { "register": "FP", "offset": -6 } },
			"dst": { "register": "AP", "offset": 1 }
		} 
	}`)

	var hint Hint
	err := json.Unmarshal(testData, &hint)
	assert.NoError(t, err)

	h, ok := hint.Args.(*TestLessThanOrEqual)
	assert.True(t, ok)

	assert.NoError(t, v.Struct(h.Lhs))
}

func TestParseHintWithBinOp(t *testing.T) {
	v := validator.New()

	testData := []byte(`{
		"TestLessThanOrEqual": {
			"lhs": { 
				"BinOp": { 
					"op": "Add", 
					"a":{ "register": "AP", "offset":1}, 
					"b":{ "Deref":{ "register": "AP", "offset":1 } } }
			},
			"rhs": { "Deref": { "register": "FP", "offset": -6 } },
			"dst": { "register": "AP", "offset": 1 }
		} 
	}`)

	var hint Hint
	err := json.Unmarshal(testData, &hint)
	assert.NoError(t, err)

	h, ok := hint.Args.(*TestLessThanOrEqual)
	assert.True(t, ok)

	assert.NoError(t, v.Struct(h))
}

func TestParsingStarknetHint(t *testing.T) {
	v := validator.New()

	testData := []byte(`{
		"SystemCall": {
            "system": {
              "Deref": {
                "register": "FP",
                "offset": -3
              }
            }
        }
	}`)

	var hint Hint
	err := json.Unmarshal(testData, &hint)
	assert.NoError(t, err)

	h, ok := hint.Args.(*SystemCall)
	assert.True(t, ok)

	assert.NoError(t, v.Struct(h))
}

func TestParsingDeprecatedHint(t *testing.T) {
	v := validator.New()

	testData := []byte(`{
		"Felt252DictRead": {
			"dict_ptr": {
				"DoubleDeref": [ { "register": "AP", "offset": 1 }, 1 ] 
			},
			"key": { "Deref": { "register": "AP", "offset": 1 } },
			"value_dst": { "register": "AP", "offset": 1 }
		}
	}`)

	var hint Hint
	err := json.Unmarshal(testData, &hint)
	assert.NoError(t, err)

	h, ok := hint.Args.(*Felt252DictRead)
	assert.True(t, ok)

	assert.NoError(t, v.Struct(h))
}
