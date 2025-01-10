package starknet

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type CairoFuncArgs struct {
	Single *fp.Element
	Array  []fp.Element
}

func ParseCairoProgramArgs(input string) ([]CairoFuncArgs, error) {
	re := regexp.MustCompile(`\[[^\]]*\]|\S+`)
	tokens := re.FindAllString(input, -1)
	var result []CairoFuncArgs

	parseValueToFelt := func(token string) (*fp.Element, error) {
		felt, err := new(fp.Element).SetString(token)
		if err != nil {
			return nil, fmt.Errorf("invalid felt value: %v", err)
		}
		return felt, nil
	}

	for _, token := range tokens {
		if single, err := parseValueToFelt(token); err == nil {
			result = append(result, CairoFuncArgs{
				Single: single,
				Array:  nil,
			})
		} else if strings.HasPrefix(token, "[") && strings.HasSuffix(token, "]") {
			arrayStr := strings.Trim(token, "[]")
			arrayElements := strings.Fields(arrayStr)
			array := make([]fp.Element, len(arrayElements))
			for i, element := range arrayElements {
				single, err := parseValueToFelt(element)
				if err != nil {
					return nil, fmt.Errorf("invalid felt value in array: %v", err)
				}
				array[i] = *single
			}
			result = append(result, CairoFuncArgs{
				Single: nil,
				Array:  array,
			})
		} else {
			return nil, fmt.Errorf("invalid token: %s", token)
		}
	}

	return result, nil
}
