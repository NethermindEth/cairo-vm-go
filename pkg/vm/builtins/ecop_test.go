package builtins

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func TestEcOp(t *testing.T) {
	wayOne()

	wayTwo()

	wayThree()

	// two and three
	two := fp.NewElement(2)
	fmt.Printf("two %d\n", two)

	three := fp.NewElement(3)
	fmt.Printf("three %d\n", three)
}

func wayOne() fp.Element {
	betaLow := &fp.Element{}
	betaLow, err := betaLow.SetString("0x609ad26c15c915c1f4cdfcb99cee9e89")
	if err != nil {
		panic(err)
	}

	betaHigh := &fp.Element{}
	betaHigh, err = betaHigh.SetString("0x6f21413efbe40de150e596d72f7a8c5")
	if err != nil {
		panic(err)
	}

	fmt.Println("beta low", betaLow)
	fmt.Println("beta high", betaHigh)

	lowBytes := betaLow.Bytes()
	highBytes := betaHigh.Bytes()

	fmt.Println("beta low bytes", lowBytes)
	fmt.Println("beta high bytes", highBytes)

	copy(lowBytes[16:32], highBytes[16:32])

	fmt.Println("beta bytes", lowBytes)

	beta, err := fp.BigEndian.Element(&lowBytes)
	if err != nil {
		panic(err)
	}

	fmt.Println("beta", beta.Text(10), beta)
	return beta
}

func wayTwo() fp.Element {
	betaLow := &fp.Element{}
	betaLow, err := betaLow.SetString("0x609ad26c15c915c1f4cdfcb99cee9e89")
	if err != nil {
		panic(err)
	}

	betaHigh := &fp.Element{}
	betaHigh, err = betaHigh.SetString("0x6f21413efbe40de150e596d72f7a8c5")
	if err != nil {
		panic(err)
	}

	for i := 0; i < 128; i++ {
		betaHigh.Double(betaHigh)
	}
	betaHigh.Add(betaHigh, betaLow)

	fmt.Println("beta 2:", betaHigh.Text(10), betaHigh)

	return *betaHigh
}

func wayThree() fp.Element {
	beta := &fp.Element{}
	beta.SetString("3141592653589793238462643383279502884197169399375105820974944592307816406665")
	fmt.Printf("beta 3: %d\n", beta)
	fmt.Printf("beta 3: %s\n", beta)
	return *beta
}
