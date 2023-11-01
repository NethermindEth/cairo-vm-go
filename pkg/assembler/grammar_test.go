package assembler

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssertEqualWithRegisterGrammar(t *testing.T) {
	code := "[fp + 3] = [ap + 4];"

	casmAst, err := parseCode(code)
	require.NoError(t, err)

	require.Equal(
		t,
		&CasmProgram{
			[]InstructionNode{
				{
					AssertEq: &AssertEq{
						Dst: &Deref{
							Name: "fp",
							Offset: &Offset{
								Sign:  "+",
								Value: ptrOf(3),
							},
						},
						Value: &Expression{
							Deref: &Deref{
								Name: "ap",
								Offset: &Offset{
									Sign:  "+",
									Value: ptrOf(4),
								},
							},
						},
					},
					ApPlusOne: false,
				},
			},
		},
		casmAst,
	)
}

func TestAssertEqualWithApPlusGrammar(t *testing.T) {
	code := "[fp + 3] = [ap + 4], ap++;"

	casmAst, err := parseCode(code)
	require.NoError(t, err)

	require.Equal(
		t,
		&CasmProgram{
			[]InstructionNode{
				{
					AssertEq: &AssertEq{
						Dst: &Deref{
							Name: "fp",
							Offset: &Offset{
								Sign:  "+",
								Value: ptrOf(3),
							},
						},
						Value: &Expression{
							Deref: &Deref{
								Name: "ap",
								Offset: &Offset{
									Sign:  "+",
									Value: ptrOf(4),
								},
							},
						},
					},
					ApPlusOne: true,
				},
			},
		},
		casmAst,
	)
}

func TestAssertEqualWithImmediateGrammar(t *testing.T) {
	code := "[fp + 1] = 5;"

	casmAst, err := parseCode(code)
	require.NoError(t, err)

	require.Equal(
		t,
		&CasmProgram{
			[]InstructionNode{
				{
					AssertEq: &AssertEq{
						Dst: &Deref{
							Name: "fp",
							Offset: &Offset{
								Sign:  "+",
								Value: ptrOf(1),
							},
						},
						Value: &Expression{
							Immediate: &ImmediateValue{
								Sign:  "",
								Value: ptrOf("5"),
							},
						},
					},
					ApPlusOne: false,
				},
			},
		},
		casmAst,
	)
}

func TestAssertEqualWithMathOperationGrammar(t *testing.T) {
	code := "[ap] = [fp + 7] + 5;"

	casmAst, err := parseCode(code)
	require.NoError(t, err)

	require.Equal(
		t,
		&CasmProgram{
			[]InstructionNode{
				{
					AssertEq: &AssertEq{
						Dst: &Deref{
							Name:   "ap",
							Offset: nil,
						},
						Value: &Expression{
							MathOperation: &MathOperation{
								Lhs: &Deref{
									Name: "fp",
									Offset: &Offset{
										Sign:  "+",
										Value: ptrOf(7),
									},
								},
								Operator: "+",
								Rhs: &DerefOrImm{
									Immediate: ptrOf("5"),
								},
							},
						},
					},
					ApPlusOne: false,
				},
			},
		},
		casmAst,
	)
}

func TestCallAbsGrammar(t *testing.T) {
	code := "call abs 123;"

	casmAst, err := parseCode(code)
	require.NoError(t, err)

	require.Equal(
		t,
		&CasmProgram{
			[]InstructionNode{
				{
					Call: &Call{
						CallType: "abs",
						Value: &DerefOrImm{
							Immediate: ptrOf("123"),
						},
					},
					ApPlusOne: false,
				},
			},
		},
		casmAst,
	)
}

func TestRetGrammar(t *testing.T) {
	code := "ret;"

	casmAst, err := parseCode(code)
	require.NoError(t, err)

	require.Equal(
		t,
		&CasmProgram{
			[]InstructionNode{
				{
					Ret: &Ret{
						Ret: "",
					},
					ApPlusOne: false,
				},
			},
		},
		casmAst,
	)
}

func TestJumpGrammar(t *testing.T) {
	for _, jmpType := range []string{"abs", "rel"} {
		t.Logf("jmpType: %s", jmpType)

		code := fmt.Sprintf("jmp %s [ap + 1] + [fp - 7];", jmpType)

		casmAst, err := parseCode(code)
		require.NoError(t, err)

		require.Equal(
			t,
			&CasmProgram{
				[]InstructionNode{
					{
						Jump: &Jump{
							JumpType: jmpType,
							Value: &Expression{
								MathOperation: &MathOperation{
									Lhs: &Deref{
										Name: "ap",
										Offset: &Offset{
											Sign:  "+",
											Value: ptrOf(1),
										},
									},
									Rhs: &DerefOrImm{
										Deref: &Deref{
											Name: "fp",
											Offset: &Offset{
												Sign:  "-",
												Value: ptrOf(7),
											},
										},
									},
									Operator: "+",
								},
							},
						},
					},
				},
			},
			casmAst,
		)
	}
}

func ptrOf[T any](n T) *T {
	return &n
}

func parseCode(code string) (*CasmProgram, error) {
	return parser.ParseString("", code)
}
