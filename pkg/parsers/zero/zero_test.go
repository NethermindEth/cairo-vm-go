package zero

import (
	starknetParser "github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrime(t *testing.T) {
	content := []byte(`
        {
            "prime": "191"
        } 
    `)

	zeroProgram, err := ZeroProgramFromJSON(content)
	require.NoError(t, err)

	require.Equal(t,
		&ZeroProgram{
			Prime: "191",
		},
		zeroProgram,
	)
}

func TestData(t *testing.T) {
	content := []byte(`
        {
            "data": [
                "0x00000002",
                "0x00000003",
                "0x00000005",
                "0x00000007"
            ]
        }
    `)
	zeroProgram, err := ZeroProgramFromJSON(content)
	require.NoError(t, err)

	require.Equal(t,
		&ZeroProgram{
			Data: []string{
				"0x00000002",
				"0x00000003",
				"0x00000005",
				"0x00000007",
			},
		},
		zeroProgram,
	)
}

func TestBuiltins(t *testing.T) {
	content := []byte(`
        {
            "builtins": [
                "output",
                "range_check",
                "bitwise"
            ]     
        } 
    `)

	zeroProgram, err := ZeroProgramFromJSON(content)
	require.NoError(t, err)

	require.Equal(t,
		&ZeroProgram{
			Builtins: []starknetParser.Builtin{
				starknetParser.Output,
				starknetParser.RangeCheck,
				starknetParser.Bitwise,
			},
		},
		zeroProgram,
	)
}

func TestHints(t *testing.T) {
	content := []byte(`
        {
            "hints": {
                "6": [
                    {
                        "accessible_scopes": [
                            "starkware.cairo.common.alloc",
                            "starkware.cairo.common.alloc.alloc"
                        ],
                        "code": "memory[ap] = segments.add()",
                        "flow_tracking_data": {
                            "ap_tracking": {
                                "group": 2,
                                "offset": 0
                            },
                            "reference_ids": {
                                "starkware.cairo.common.math.assert_nn.a": 9,
                                "starkware.cairo.common.math.assert_nn.range_check_ptr": 10
                            }
                        }
                    }
                ]
            }
        }
    `)

	zeroProgram, err := ZeroProgramFromJSON(content)
	require.NoError(t, err)

	require.Equal(t,
		&ZeroProgram{
			Hints: map[string][]Hint{
				"6": {
					{
						AccessibleScopes: []string{
							"starkware.cairo.common.alloc",
							"starkware.cairo.common.alloc.alloc",
						},
						Code: "memory[ap] = segments.add()",
						FlowTrackingData: FlowTrackingData{
							ApTracking: ApTracking{
								Group:  2,
								Offset: 0,
							},
							ReferenceIds: map[string]uint64{
								"starkware.cairo.common.math.assert_nn.a":               9,
								"starkware.cairo.common.math.assert_nn.range_check_ptr": 10,
							},
						},
					},
				},
			},
		},
		zeroProgram,
	)
}

func TestCompilerVersion(t *testing.T) {
	content := []byte(`
        {
            "prime": "0x05"
        }
    `)
	zeroProgram, err := ZeroProgramFromJSON(content)
	require.NoError(t, err)

	require.Equal(
		t,
		&ZeroProgram{
			Prime: "0x05",
		},
		zeroProgram,
	)
}

func TestMainScope(t *testing.T) {
	content := []byte(`
        {
            "main_scope": "__main__"
        }
    `)
	zeroProgram, err := ZeroProgramFromJSON(content)
	require.NoError(t, err)

	require.Equal(
		t,
		&ZeroProgram{
			MainScope: "__main__",
		},
		zeroProgram,
	)
}

func TestIdentifiers(t *testing.T) {
	content := []byte(`
        {
            "identifiers": {
                "__main__.BitwiseBuiltin": {
                    "destination": "starkware.cairo.common.cairo_builtins.BitwiseBuiltin",
                    "type": "alias"
                },
                "__main__.fill_array.Args": {
                    "full_name": "__main__.fill_array.Args",
                    "members": {
                        "array": {
                            "cairo_type": "felt*",
                            "offset": 0
                        }
                    },
                    "size": 1,
                    "type": "struct"
                },
                "__main__.fill_array.__temp18": {
                    "cairo_type": "felt",
                    "full_name": "__main__.fill_array.__temp18",
                    "references": [
                        {
                            "ap_tracking_data": {
                                "group": 26,
                                "offset": 1
                            },
                            "pc": 312,
                            "value": "[cast(ap + (-1), felt*)]"
                        }
                    ],
                    "type": "reference"
                }
            }
        }
    `)
	zeroProgram, err := ZeroProgramFromJSON(content)
	require.NoError(t, err)

	require.Equal(
		t,
		&ZeroProgram{
			Identifiers: map[string]any{
				"__main__.BitwiseBuiltin": map[string]any{
					"destination": "starkware.cairo.common.cairo_builtins.BitwiseBuiltin",
					"type":        "alias",
				},
				"__main__.fill_array.Args": map[string]any{
					"full_name": "__main__.fill_array.Args",
					"members": map[string]any{
						"array": map[string]any{
							"cairo_type": "felt*",
							"offset":     float64(0),
						},
					},
					"size": float64(1),
					"type": "struct",
				},
				"__main__.fill_array.__temp18": map[string]any{
					"cairo_type": "felt",
					"full_name":  "__main__.fill_array.__temp18",
					"references": []any{
						map[string]any{
							"ap_tracking_data": map[string]any{
								"group":  float64(26),
								"offset": float64(1),
							},
							"pc":    float64(312),
							"value": "[cast(ap + (-1), felt*)]",
						},
					},
					"type": "reference",
				},
			},
		},
		zeroProgram,
	)
}

func TestAtributes(t *testing.T) {
	content := []byte(`
        {
            "attributes": [
                {
                    "start_pc": 13,
                    "end_pc": 17,
                    "flow_tracking_data": {
                        "ap_tracking": {
                            "group": 2,
                            "offset": 0
                        },
                        "reference_ids": {
                            "ref_1": 1,
                            "ref_2": 2
                        }
                    },
                    "accessible_scopes": ["scope1", "scope2"]
                }
            ]
        }
    `)

	zeroProgram, err := ZeroProgramFromJSON(content)
	require.NoError(t, err)

	require.Equal(
		t,
		&ZeroProgram{
			Attributes: []AttributeScope{
				{
					StartPc: 13,
					EndPc:   17,
					FlowTrackingData: FlowTrackingData{
						ApTracking: ApTracking{
							Group:  2,
							Offset: 0,
						},
						ReferenceIds: map[string]uint64{
							"ref_1": 1,
							"ref_2": 2,
						},
					},
					AccessibleScopes: []string{
						"scope1",
						"scope2",
					},
				},
			},
		},
		zeroProgram,
	)
}

func TestDebugInfo(t *testing.T) {
	content := []byte(`
        {
            "debug_info": {
                "file_contents": {
                    "key1": "value1"
                },
                "instruction_locations": {
                    "7": {
                        "accessible_scopes": [
                            "starkware.cairo.lang.compiler.lib.registers",
                            "starkware.cairo.lang.compiler.lib.registers.get_fp_and_pc"
                        ],
                        "flow_tracking_data": {
                            "ap_tracking": {
                                "group": 0,
                                "offset": 0
                            },
                            "reference_ids": {
                                "starkware.cairo.lang.compiler.lib.registers.get_ap.fp_val": 0,
                                "starkware.cairo.lang.compiler.lib.registers.get_ap.pc_val": 1
                            }
                        },
                        "hints": [
                            {
                                "location": {
                                    "end_col": 38,
                                    "end_line": 3,
                                    "input_file": {
                                        "filename": "some/path"
                                    },
                                    "start_col": 5,
                                    "start_line": 3
                                },
                                "n_prefix_newlines": 0
                            }
                        ],
                        "inst": {
                            "end_col": 73,
                            "end_line": 7,
                            "input_file": {
                                "filename": "some/path"
                            },
                            "start_col": 5,
                            "start_line": 7
                        }
                    }
                }
            }
        }
    `)

	zeroProgram, err := ZeroProgramFromJSON(content)
	require.NoError(t, err)

	require.Equal(
		t,
		&ZeroProgram{
			DebugInfo: DebugInfo{
				FileContents: map[string]string{
					"key1": "value1",
				},
				InstructionLocations: map[string]InstructionLocation{
					"7": {
						AccessibleScopes: []string{
							"starkware.cairo.lang.compiler.lib.registers",
							"starkware.cairo.lang.compiler.lib.registers.get_fp_and_pc",
						},
						FlowTrackingData: FlowTrackingData{
							ApTracking: ApTracking{
								Group:  0,
								Offset: 0,
							},
							ReferenceIds: map[string]uint64{
								"starkware.cairo.lang.compiler.lib.registers.get_ap.fp_val": 0,
								"starkware.cairo.lang.compiler.lib.registers.get_ap.pc_val": 1,
							},
						},
						Hints: []HintLocation{
							{
								Location: Location{
									EndCol:  38,
									EndLine: 3,
									InputFile: map[string]string{
										"filename": "some/path",
									},
									StartCol:  5,
									StartLine: 3,
								},
								NPrefixNewlines: 0,
							},
						},
						Inst: Location{
							EndCol:  73,
							EndLine: 7,
							InputFile: map[string]string{
								"filename": "some/path",
							},
							StartCol:  5,
							StartLine: 7,
						},
					},
				},
			},
		},
		zeroProgram,
	)
}
