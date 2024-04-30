package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestZeroHintDictionaries(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"SquashDictInnerAssertKeys": {
			{operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.ScopeManager.EnterScope(map[string]any{"keys": ctx.SquashedDictionaryManager.Keys})
					err := hinter.InitializeSquashedDictionaryManager(ctx)
					if err != nil {
						t.Fatal(err.Error())
					}
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerNextKeyHint()
				},
				check: func(t *testing.T, ctx *hintTestContext) {},
			},
			{
				operanders: []*hintOperander{},
				ctxInit: func(ctx *hinter.HintRunnerContext) {
					ctx.SquashedDictionaryManager.Keys = append(ctx.SquashedDictionaryManager.Keys, fp.Element{})
					ctx.ScopeManager.EnterScope(map[string]any{"keys": ctx.SquashedDictionaryManager.Keys})
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newSquashDictInnerNextKeyHint()
				},
				errCheck: errorTextContains("len(keys) == 0` No keys left for processing"),
			},
		},
	})
}
