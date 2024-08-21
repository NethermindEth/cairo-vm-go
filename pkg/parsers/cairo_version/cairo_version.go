package cairoversion

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

type CairoVersion struct {
	Version string `json:"compiler_version"`
}

func GetCairoVersion(pathToFile string) (uint8, error) {
	content, err := os.ReadFile(pathToFile)
	if err != nil {
		return 0, err
	}
	cv := CairoVersion{}
	err = json.Unmarshal(content, &cv)
	if err != nil {
		return 0, err
	}
	firstNumberStr := strings.Split(cv.Version, ".")[0]
	firstNumber, err := strconv.ParseUint(firstNumberStr, 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(firstNumber), nil
}
