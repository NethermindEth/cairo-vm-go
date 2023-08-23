package zero

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZeroParse(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected ZeroProgram
	}{
		{
			name: "Attributes field",
			jsonData: `
				{
					"attributes": ["attr1", "attr2"]
				}`,
			expected: ZeroProgram{
				Attributes: []string{"attr1", "attr2"},
			},
		},
		{
			name: "Builtins field",
			jsonData: `
				{
					"builtins": {"builtin1": 1, "builtin2": 2}
				}`,
			expected: ZeroProgram{
				Builtins: map[string]int64{"builtin1": 1, "builtin2": 2},
			},
		},
		{
			name: "Code field",
			jsonData: `
				{
					"code": "some code sample"
				}`,
			expected: ZeroProgram{
				Code: "some code sample",
			},
		},
		{
			name: "Version field",
			jsonData: `
				{
					"version": "0.12.2"
				}`,
			expected: ZeroProgram{
				Version: "0.12.2",
			},
		},
		{
			name: "Prime field",
			jsonData: `
				{
					"prime": "0x800000000000011000000000000000000000000000000000000000000000001"
				}`,
			expected: ZeroProgram{
				Prime: "0x800000000000011000000000000000000000000000000000000000000000001",
			},
		},
		{
			name: "Hints field",
			jsonData: `
				{
					"hints": {
						"0": [
							{
								"accessible_scopes": [
									"starkware.cairo.common.math",
									"starkware.cairo.common.math.assert_not_equal"
								],
								"code": "from starkware.cairo.lang.vm.relocatable import RelocatableValue\nboth_ints = isinstance(ids.a, int) and isinstance(ids.b, int)\nboth_relocatable = (\n    isinstance(ids.a, RelocatableValue) and isinstance(ids.b, RelocatableValue) and\n    ids.a.segment_index == ids.b.segment_index)\nassert both_ints or both_relocatable, \\\n    f'assert_not_equal failed: non-comparable values: {ids.a}, {ids.b}.'\nassert (ids.a - ids.b) % PRIME != 0, f'assert_not_equal failed: {ids.a} = {ids.b}.'"
							}
						]
					}
				}`,
			expected: ZeroProgram{
				Hints: map[string][]Hint{
					"0": {{
						AccessibleScopes: []string{"starkware.cairo.common.math", "starkware.cairo.common.math.assert_not_equal"},
						Code:             "from starkware.cairo.lang.vm.relocatable import RelocatableValue\nboth_ints = isinstance(ids.a, int) and isinstance(ids.b, int)\nboth_relocatable = (\n    isinstance(ids.a, RelocatableValue) and isinstance(ids.b, RelocatableValue) and\n    ids.a.segment_index == ids.b.segment_index)\nassert both_ints or both_relocatable, \\\n    f'assert_not_equal failed: non-comparable values: {ids.a}, {ids.b}.'\nassert (ids.a - ids.b) % PRIME != 0, f'assert_not_equal failed: {ids.a} = {ids.b}.'",
					}},
				},
			},
		},
		// TODO: Add debug-info test case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var z ZeroProgram
			err := json.Unmarshal([]byte(tt.jsonData), &z)
			if err != nil {
				t.Errorf("Failed to unmarshal JSON: %s", err)
			}

			assert.Equal(t, tt.expected, z, "Field does not match")
		})
	}
}
