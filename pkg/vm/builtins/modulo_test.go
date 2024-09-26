package builtins

import (
	"math/big"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/require"
)

func TestModuloBuiltin(t *testing.T) {
	mod := &ModBuiltin{ratio: 128, wordBitLen: 1, batchSize: 96, modBuiltinType: Add}
	segment := memory.EmptySegmentWithLength(9)
	segment.WithBuiltinRunner(mod)

	v0 := new(fp.Element).SetUint64(1)
	v1 := new(fp.Element).SetUint64(2)
	v2 := new(fp.Element).SetUint64(3)
	v3 := new(fp.Element).SetUint64(4)
	v4 := new(fp.Element).SetUint64(5)
	v5 := new(fp.Element).SetUint64(9)
	v6 := new(fp.Element).SetUint64(7)
	v7 := new(fp.Element).SetUint64(8)
	v00 := memory.MemoryValueFromFieldElement(v0)
	v11 := memory.MemoryValueFromFieldElement(v1)
	v22 := memory.MemoryValueFromFieldElement(v2)
	v33 := memory.MemoryValueFromFieldElement(v3)
	v44 := memory.MemoryValueFromFieldElement(v4)
	v55 := memory.MemoryValueFromFieldElement(v5)
	v66 := memory.MemoryValueFromFieldElement(v6)
	v77 := memory.MemoryValueFromFieldElement(v7)
	require.NoError(t, segment.Write(0, &v00))
	require.NoError(t, segment.Write(1, &v11))
	require.NoError(t, segment.Write(2, &v22))
	require.NoError(t, segment.Write(3, &v33))
	require.NoError(t, segment.Write(4, &v44))
	require.NoError(t, segment.Write(5, &v55))
	require.NoError(t, segment.Write(6, &v66))
	require.NoError(t, segment.Write(7, &v77))

	// TODO: Once Checkwrite and InferValue are implemented, uncomment this
	// k, err := segment.Read(9)
	// require.NoError(t, err)
	// fmt.Println(k)
	// ans, err := k.FieldElement()
	// require.NoError(t, err)
	// expected := fp.NewElement(0)
	// _, err = expected.SetUint64("15")
	// require.NoError(t, err)
	// assert.Equal(t, ans, &expected)
}

/*
Tests whether runner completes a trio a, b, c as the input implies:
If inverse_bool is False it tests whether a=x1, b=x2, c=None will be completed with c=res.
If inverse_bool is True it tests whether c=x1, b=x2, a=None will be completed with a=res.
The case c=x1, a=x2, b=None is currently completely symmetric in fill_value so it isn't tested.
*/
func checkResult(runner ModBuiltin, inverse bool, p, x1, x2 big.Int) (*big.Int, error) {
	mem := memory.Memory{}

	mem.AllocateEmptySegment()

	offsetsPtr := memory.MemoryAddress{SegmentIndex: 0, Offset: 0}

	for i := 0; i < 3; i++ {
		offsetsPtrAddr, err := offsetsPtr.AddOffset(int16(i))
		if err != nil {
			return nil, err
		}

		mv := memory.MemoryValueFromInt(i * N_WORDS)
		if err := mem.WriteToAddress(&offsetsPtrAddr, &mv); err != nil {
			return nil, err
		}
	}
	x1Addr := memory.MemoryAddress{SegmentIndex: 0, Offset: 24}
	x2Addr, err := x1Addr.AddOffset(int16(N_WORDS))
	if err != nil {
		return nil, err
	}

	runner.writeNWordsValue(&mem, x2Addr, x2)
	resAddr, err := x1Addr.AddOffset(int16(2 * N_WORDS))

	if err != nil {
		return nil, err
	}

	if inverse {
		x1Addr, resAddr = resAddr, x1Addr
	}

	runner.writeNWordsValue(&mem, x1Addr, x1)

	runner.fillValue(&mem, ModBuiltinInputs{
		p:          p,
		pValues:    [N_WORDS]fp.Element{}, // not used in fillValue
		valuesPtr:  x1Addr,
		n:          0, // not used in fillValue
		offsetsPtr: offsetsPtr,
	}, 0, Operation(runner.modBuiltinType), Operation("Inv"+runner.modBuiltinType))

	_, OutRes, err := runner.readNWordsValue(&mem, resAddr)
	if err != nil {
		return nil, err
	}

	return &OutRes, nil
}

func TestAddModBuiltinRunnerAddition(t *testing.T) {
	runner := NewModBuiltin(128, 1, 96, Add)
	res1, err := checkResult(*runner, false, *big.NewInt(67), *big.NewInt(17), *big.NewInt(40))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(57), res1)
	res2, err := checkResult(*runner, false, *big.NewInt(67), *big.NewInt(82), *big.NewInt(31))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(46), res2)
	res3, err := checkResult(*runner, false, *big.NewInt(67), *big.NewInt(68), *big.NewInt(69))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(70), res3)
	res4, err := checkResult(*runner, false, *big.NewInt(67), *big.NewInt(68), *big.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(1), res4)
}
