package utils

import (
	"errors"
	"math/big"
)

func GetCairoPrime() (big.Int, bool) {
	// 2**251 + 17 * 2**192 + 1
	cairoPrime, ok := new(big.Int).SetString("3618502788666131213697322783095070105623107215331596699973092056135872020481", 10)
	return *cairoPrime, ok
}

func getFeltMaxHalved() (big.Int, bool) {
	feltMaxHalved, ok := new(big.Int).SetString("1809251394333065606848661391547535052811553607665798349986546028067936010240", 10)
	return *feltMaxHalved, ok
}

func isQuadResidue(n, fieldPrime *big.Int) bool {
	if n.Cmp(big.NewInt(0)) == 0 || n.Cmp(big.NewInt(1)) == 0 {
		return true
	}
	feltMaxHalved, _ := getFeltMaxHalved()
	modPowResult := new(big.Int).Exp(n, &feltMaxHalved, fieldPrime)
	return modPowResult.Cmp(big.NewInt(1)) == 0
}

func ySquaredFromX(x, alpha, beta, fieldPrime *big.Int) *big.Int {
	// y^2 = (x^3 + alpha * x + beta) % fieldPrime
	ySquared := new(big.Int).Exp(x, big.NewInt(3), fieldPrime)
	ySquared.Add(ySquared, new(big.Int).Mul(alpha, x))
	ySquared.Add(ySquared, beta)
	ySquared.Mod(ySquared, fieldPrime)
	return ySquared
}

func RecoverY(x, alpha, beta, fieldPrime *big.Int) (*big.Int, error) {
	ySquared := ySquaredFromX(x, alpha, beta, fieldPrime)
	if isQuadResidue(ySquared, fieldPrime) {
		return sqrtMod(ySquared, fieldPrime), nil
	}
	return nil, errors.New("not on curve")
}

func sqrtMod(n, p *big.Int) *big.Int {
	root1 := new(big.Int).ModSqrt(n, p)
	if root1 == nil {
		return nil
	}
	root2 := new(big.Int).Sub(p, root1)
	if root1.Cmp(root2) < 0 {
		return root1
	}
	return root2
}
