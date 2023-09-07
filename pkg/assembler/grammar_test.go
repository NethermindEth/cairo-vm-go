package assembler

import (
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/stretchr/testify/require"
)

var testParser *participle.Parser[CasmProgram] = safeParserBuild()

func TestAssertEqualWithRegister(t *testing.T) {
	code := "[fp + 3] = [ap + 4];"

	casmAst, err := parseCode(code)
	require.NoError(t, err)

	require.Equal(
		t,
		&CasmProgram{
			[]Instruction{
				{
					Core: &CoreInstruction{
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
					},
					ApPlusOne: false,
				},
			},
		},
		casmAst,
	)
}

func ptrOf[T any](n T) *T {
	return &n
}

func parseCode(code string) (*CasmProgram, error) {
	return testParser.ParseString("", code)
}

func safeParserBuild() *participle.Parser[CasmProgram] {
	parser, err := participle.Build[CasmProgram]()
	if err != nil {
		panic(err)
	}
	return parser
}
