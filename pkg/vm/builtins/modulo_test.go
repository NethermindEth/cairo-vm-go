package builtins

import (
	// "fmt"
	"math/big"
	"testing"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/stretchr/testify/require"
)

/*
Tests whether runner completes a trio a, b, c as the input implies:
If inverse is False it tests whether a = x1, b = x2, c = None will be completed with c = res.
If inverse is True it tests whether c = x1, b = x2, a = None will be completed with a = res.
*/
func checkResult(runner ModBuiltin, inverse bool, p, x1, x2 big.Int) (*big.Int, error) {
	mem := memory.Memory{}

	mem.AllocateBuiltinSegment(&runner)

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

	valuesAddr := memory.MemoryAddress{SegmentIndex: 0, Offset: 24}

	x1Addr, err := valuesAddr.AddOffset(int16(0))
	if err != nil {
		return nil, err
	}

	x2Addr, err := valuesAddr.AddOffset(int16(N_WORDS))
	if err != nil {
		return nil, err
	}
	err = runner.writeNWordsValue(&mem, x2Addr, x2)
	if err != nil {
		return nil, err
	}

	resAddr, err := valuesAddr.AddOffset(int16(2 * N_WORDS))
	if err != nil {
		return nil, err
	}

	if inverse {
		x1Addr, resAddr = resAddr, x1Addr
	}

	err = runner.writeNWordsValue(&mem, x1Addr, x1)
	if err != nil {
		return nil, err
	}

	_, err = runner.fillValue(&mem, ModBuiltinInputs{
		p:          p,
		pValues:    [N_WORDS]fp.Element{}, // not used in fillValue
		valuesPtr:  valuesAddr,
		n:          0, // not used in fillValue
		offsetsPtr: offsetsPtr,
	}, 0, runner.modBuiltinType)

	if err != nil {
		return nil, err
	}

	_, OutRes, err := runner.readNWordsValue(&mem, resAddr)
	if err != nil {
		return nil, err
	}

	return OutRes, nil
}

func TestAddModBuiltinRunnerAddition(t *testing.T) {
	runner := NewModBuiltin(1, 3, 1, Add)
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
	_, err = checkResult(*runner, false, *big.NewInt(4094), *big.NewInt(4095), *big.NewInt(4095))
	require.ErrorContains(t, err, "Expected a Add b - 1 * p <= 4095")
}

func TestAddModBuiltinRunnerSubtraction(t *testing.T) {
	runner := NewModBuiltin(1, 3, 1, Add)
	res1, err := checkResult(*runner, true, *big.NewInt(67), *big.NewInt(52), *big.NewInt(38))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(14), res1)
	res2, err := checkResult(*runner, true, *big.NewInt(67), *big.NewInt(5), *big.NewInt(68))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(4), res2)
	res3, err := checkResult(*runner, true, *big.NewInt(67), *big.NewInt(5), *big.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(5), res3)
	res4, err := checkResult(*runner, true, *big.NewInt(67), *big.NewInt(0), *big.NewInt(5))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(62), res4)
	_, err = checkResult(*runner, true, *big.NewInt(67), *big.NewInt(70), *big.NewInt(138))
	require.ErrorContains(t, err, "addend greater than sum + p")
}

func TestMulModBuiltinRunnerMultiplication(t *testing.T) {
	runner := NewModBuiltin(1, 3, 1, Mul)
	res1, err := checkResult(*runner, false, *big.NewInt(67), *big.NewInt(11), *big.NewInt(8))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(21), res1)
	res2, err := checkResult(*runner, false, *big.NewInt(67), *big.NewInt(68), *big.NewInt(69))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(2), res2)
	res3, err := checkResult(*runner, false, *big.NewInt(67), *big.NewInt(525), *big.NewInt(526))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(1785), res3)
	res4, err := checkResult(*runner, false, *big.NewInt(67), *big.NewInt(525), *big.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(0), res4)
	_, err = checkResult(*runner, false, *big.NewInt(67), *big.NewInt(3777), *big.NewInt(3989))
	require.ErrorContains(t, err, "Expected a Mul b - 4095 * p <= 4095")
}

func TestMulModBuiltinRunnerDivision(t *testing.T) {
	runner := NewModBuiltin(1, 3, 1, Mul)
	res1, err := checkResult(*runner, true, *big.NewInt(67), *big.NewInt(36), *big.NewInt(9))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(4), res1)
	res2, err := checkResult(*runner, true, *big.NewInt(67), *big.NewInt(138), *big.NewInt(41))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(5), res2)
	res3, err := checkResult(*runner, true, *big.NewInt(67), *big.NewInt(272), *big.NewInt(41))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(72), res3)
	res4, err := checkResult(*runner, true, *big.NewInt(67), *big.NewInt(0), *big.NewInt(0))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(1), res4)
	res5, err := checkResult(*runner, true, *big.NewInt(66), *big.NewInt(6), *big.NewInt(3))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(22), res5)
}
