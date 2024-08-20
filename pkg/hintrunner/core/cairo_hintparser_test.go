package core

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/require"
)

func TestCairoHintParser(t *testing.T) {
	type testSetType struct {
		CairoProgramString []byte
		ExpectedHintMap    map[uint64][]hinter.Hinter
	}

	testSet := []testSetType{
		{
			CairoProgramString: []byte(`
				{
					"hints": [
						[
							2,
							[
								{
									"AllocFelt252Dict": {
										"segment_arena_ptr": {
											"Deref": {
												"register": "FP",
												"offset": -4
											}
										}
									}
								}
							]
						],
						[
							13,
							[
								{
									"AllocFelt252Dict": {
										"segment_arena_ptr": {
											"BinOp": {
												"op": "Add",
												"a": {
													"register": "FP",
													"offset": -4
												},
												"b": {
													"Immediate": "0x3"
												}
											}
										}
									}
								}
							]
						],
						[
							46,
							[
								{
									"Felt252DictEntryInit": {
										"dict_ptr": {
											"Deref": {
												"register": "FP",
												"offset": 2
											}
										},
										"key": {
											"Deref": {
												"register": "AP",
												"offset": -1
											}
										}
									}
								}
							]
						],
						[
							49,
							[
								{
									"Felt252DictEntryUpdate": {
										"dict_ptr": {
											"BinOp": {
												"op": "Add",
												"a": {
													"register": "FP",
													"offset": 2
												},
												"b": {
													"Immediate": "0x3"
												}
											}
										},
										"value": {
											"Deref": {
												"register": "AP",
												"offset": -1
											}
										}
									}
								}
							]
						],
						[
							104,
							[
								{
									"GetSegmentArenaIndex": {
										"dict_end_ptr": {
											"Deref": {
												"register": "FP",
												"offset": -3
											}
										},
										"dict_index": {
											"register": "FP",
											"offset": 0
										}
									}
								}
							]
						],
						[
							145,
							[
								{
									"AllocSegment": {
										"dst": {
											"register": "FP",
											"offset": 3
										}
									}
								}
							]
						],
						[
							153,
							[
								{
									"InitSquashData": {
										"dict_accesses": {
											"Deref": {
												"register": "FP",
												"offset": -4
											}
										},
										"ptr_diff": {
											"Deref": {
												"register": "FP",
												"offset": 0
											}
										},
										"n_accesses": {
											"Deref": {
												"register": "AP",
												"offset": -1
											}
										},
										"big_keys": {
											"register": "FP",
											"offset": 2
										},
										"first_key": {
											"register": "FP",
											"offset": 1
										}
									}
								}
							]
						],
						[
							172,
							[
								{
									"GetCurrentAccessIndex": {
										"range_check_ptr": {
											"Deref": {
												"register": "FP",
												"offset": -9
											}
										}
									}
								}
							]
						],
						[
							185,
							[
								{
									"ShouldSkipSquashLoop": {
										"should_skip_loop": {
											"register": "AP",
											"offset": -4
										}
									}
								}
							]
						],
						[
							187,
							[
								{
									"GetCurrentAccessDelta": {
										"index_delta_minus1": {
											"register": "AP",
											"offset": 0
										}
									}
								}
							]
						],
						[
							198,
							[
								{
									"ShouldContinueSquashLoop": {
										"should_continue": {
											"register": "AP",
											"offset": -4
										}
									}
								}
							]
						],
						[
							212,
							[
								{
									"GetNextDictKey": {
										"next_key": {
											"register": "FP",
											"offset": 0
										}
									}
								}
							]
						],
						[
							231,
							[
								{
									"AssertLeFindSmallArcs": {
										"range_check_ptr": {
											"BinOp": {
												"op": "Add",
												"a": {
													"register": "AP",
													"offset": -4
												},
												"b": {
													"Immediate": "0x1"
												}
											}
										},
										"a": {
											"Deref": {
												"register": "FP",
												"offset": -6
											}
										},
										"b": {
											"Deref": {
												"register": "FP",
												"offset": 0
											}
										}
									}
								}
							]
						],
						[
							243,
							[
								{
									"AssertLeIsFirstArcExcluded": {
										"skip_exclude_a_flag": {
											"register": "AP",
											"offset": 0
										}
									}
								}
							]
						],
						[
							255,
							[
								{
									"AssertLeIsSecondArcExcluded": {
										"skip_exclude_b_minus_a": {
											"register": "AP",
											"offset": 0
										}
									}
								}
							]
						]
					]
				}
			`),
			ExpectedHintMap: map[uint64][]hinter.Hinter{
				2: {
					&AllocFelt252Dict{
						SegmentArenaPtr: hinter.Deref{
							Deref: hinter.FpCellRef(-4),
						},
					},
				},
				13: {
					&AllocFelt252Dict{
						SegmentArenaPtr: hinter.BinaryOp{
							Operator: hinter.Add,
							Lhs:      hinter.FpCellRef(-4),
							Rhs:      hinter.Immediate(*new(fp.Element).SetInt64(3)),
						},
					},
				},
				46: {
					&Felt252DictEntryInit{
						DictPtr: hinter.Deref{
							Deref: hinter.FpCellRef(2),
						},
						Key: hinter.Deref{
							Deref: hinter.ApCellRef(-1),
						},
					},
				},
				49: {
					&Felt252DictEntryUpdate{
						DictPtr: hinter.BinaryOp{
							Operator: hinter.Add,
							Lhs:      hinter.FpCellRef(2),
							Rhs:      hinter.Immediate(*new(fp.Element).SetInt64(3)),
						},
						Value: hinter.Deref{
							Deref: hinter.ApCellRef(-1),
						},
					},
				},
				104: {
					&GetSegmentArenaIndex{
						DictEndPtr: hinter.Deref{
							Deref: hinter.FpCellRef(-3),
						},
						DictIndex: hinter.FpCellRef(0),
					},
				},
				145: {
					&AllocSegment{
						Dst: hinter.FpCellRef(3),
					},
				},
				153: {
					&InitSquashData{
						DictAccesses: hinter.Deref{
							Deref: hinter.FpCellRef(-4),
						},
						NumAccesses: hinter.Deref{
							Deref: hinter.ApCellRef(-1),
						},
						BigKeys:  hinter.FpCellRef(2),
						FirstKey: hinter.FpCellRef(1),
					},
				},
				172: {
					&GetCurrentAccessIndex{
						RangeCheckPtr: hinter.Deref{
							Deref: hinter.FpCellRef(-9),
						},
					},
				},
				185: {
					&ShouldSkipSquashLoop{
						ShouldSkipLoop: hinter.ApCellRef(-4),
					},
				},
				187: {
					&GetCurrentAccessDelta{
						IndexDeltaMinusOne: hinter.ApCellRef(0),
					},
				},
				198: {
					&ShouldContinueSquashLoop{
						ShouldContinue: hinter.ApCellRef(-4),
					},
				},
				212: {
					&GetNextDictKey{
						NextKey: hinter.FpCellRef(0),
					},
				},
				231: {
					&AssertLeFindSmallArc{
						RangeCheckPtr: hinter.BinaryOp{
							Operator: hinter.Add,
							Lhs:      hinter.ApCellRef(-4),
							Rhs:      hinter.Immediate(*new(fp.Element).SetInt64(1)),
						},
						A: hinter.Deref{
							Deref: hinter.FpCellRef(-6),
						},
						B: hinter.Deref{
							Deref: hinter.FpCellRef(0),
						},
					},
				},
				243: {
					&AssertLeIsFirstArcExcluded{
						SkipExcludeAFlag: hinter.ApCellRef(0),
					},
				},
				255: {
					&AssertLeIsSecondArcExcluded{
						SkipExcludeBMinusA: hinter.ApCellRef(0),
					},
				},
			},
		},
	}

	for _, test := range testSet {
		starknetProgram, err := starknet.StarknetProgramFromJSON(test.CairoProgramString)
		require.NoError(t, err)
		output, err := GetCairoHints(starknetProgram)
		require.NoError(t, err)

		require.Equal(t, test.ExpectedHintMap, output, "Hint maps do not match")

	}
}
