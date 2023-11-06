package builtins

import (
	"fmt"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/require"
)

func TestECDSA(t *testing.T) {
	ecdsa := &ECDSA{}
	segment := memory.EmptySegmentWithLength(5)
	segment.WithBuiltinRunner(ecdsa)

	pubkey, _ := new(fp.Element).SetString("1839793652349538280924927302501143912227271479439798783640887258675143576352")
	msg, _ := new(fp.Element).SetString("1839793652349538280924927302501143912227271479439798783640887258675143576352")
	r, _ := new(fp.Element).SetString("1839793652349538280924927302501143912227271479439798783640887258675143576352")
	s, _ := new(fp.Element).SetString("1819432147005223164874083361865404672584671743718628757598322238853218813979")

	pubkeyValue := memory.MemoryValueFromFieldElement(pubkey)
	msgValue := memory.MemoryValueFromFieldElement(msg)

	require.NoError(t, segment.Write(0, &pubkeyValue))
	require.NoError(t, segment.Write(1, &msgValue))
	ecdsa.AddSignature(0, *r, *s)

	fmt.Println(segment.Read(2))
}
