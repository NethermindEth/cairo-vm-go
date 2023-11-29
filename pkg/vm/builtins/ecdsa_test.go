package builtins

import (
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/require"
)

func TestECDSA(t *testing.T) {
	ecdsa := &ECDSA{}
	segment := memory.EmptySegmentWithLength(5)
	segment.WithBuiltinRunner(ecdsa)

	pubkey, _ := new(fp.Element).SetString("1735102664668487605176656616876767369909409133946409161569774794110049207117")
	msg, _ := new(fp.Element).SetString("2718")
	r, _ := new(fp.Element).SetString("3086480810278599376317923499561306189851900463386393948998357832163236918254")
	s, _ := new(fp.Element).SetString("598673427589502599949712887611119751108407514580626464031881322743364689811")

	pubkeyValue := memory.MemoryValueFromFieldElement(pubkey)
	msgValue := memory.MemoryValueFromFieldElement(msg)

	require.NoError(t, ecdsa.AddSignature(0, r, s))
	require.NoError(t, segment.Write(1, &msgValue))
	require.NoError(t, segment.Write(0, &pubkeyValue))

}
func TestECDSAInvalidSig(t *testing.T) {
	ecdsa := &ECDSA{}
	segment := memory.EmptySegmentWithLength(5)
	segment.WithBuiltinRunner(ecdsa)

	pubkey, _ := new(fp.Element).SetString("1735102664668487605176656616876767369909409133946409161569774794110049207117")
	msg, _ := new(fp.Element).SetString("999999999999999")
	r, _ := new(fp.Element).SetString("4123123123213")
	s, _ := new(fp.Element).SetString("31231231313")

	pubkeyValue := memory.MemoryValueFromFieldElement(pubkey)
	msgValue := memory.MemoryValueFromFieldElement(msg)

	require.NoError(t, ecdsa.AddSignature(0, r, s))
	require.NoError(t, segment.Write(0, &pubkeyValue))
	err := segment.Write(1, &msgValue)
	require.ErrorContains(t, err, "signature is not valid")

}
