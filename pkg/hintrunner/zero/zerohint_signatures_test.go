package zero

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestSignatures(t *testing.T) {
	runHinterTests(t, map[string][]hintTestCase{
		"VerifyECDSASignature": {
			{
				operanders: []*hintOperander{
					{Name: "ecdsaPtr", Kind: reference, Value: addrBuiltin(starknet.ECDSA, 0)},
					{Name: "signature_r", Kind: apRelative, Value: feltString("3086480810278599376317923499561306189851900463386393948998357832163236918254")},
					{Name: "signature_s", Kind: apRelative, Value: feltString("598673427589502599949712887611119751108407514580626464031881322743364689811")},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					ecdsaPtr := ctx.operanders["ecdsaPtr"].(*hinter.DoubleDeref).Deref
					return newVerifyECDSASignatureHinter(ecdsaPtr, ctx.operanders["signature_r"], ctx.operanders["signature_s"])
				},
				errCheck: func(t *testing.T, ctx *hintTestContext, err error) {
					require.NoError(t, err)
				},
			},
		},
		"GetPointFromX": {
			{
				operanders: []*hintOperander{
					{Name: "xCube.d0", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "xCube.d1", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "xCube.d2", Kind: apRelative, Value: &utils.FeltZero},
					{Name: "v", Kind: apRelative, Value: &utils.FeltZero},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newGetPointFromXHinter(ctx.operanders["xCube.d0"], ctx.operanders["v"])
				},
				check: varValueInScopeEquals("value", bigIntString("64828261740814840065360381756190772627110652128289340260788836867053167272156")),
			},
			{
				operanders: []*hintOperander{
					{Name: "xCube.d0", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "xCube.d1", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "xCube.d2", Kind: apRelative, Value: &utils.FeltOne},
					{Name: "v", Kind: apRelative, Value: &utils.FeltOne},
				},
				makeHinter: func(ctx *hintTestContext) hinter.Hinter {
					return newGetPointFromXHinter(ctx.operanders["xCube.d0"], ctx.operanders["v"])
				},
				check: varValueInScopeEquals("value", bigIntString("82756063490943812239858893365668976201432184968019401309780855007018022222141")),
			},
		},
	})
}
