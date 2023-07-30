package vm

import (
	"encoding/json"
	"os"

	f "github.com/NethermindEth/juno/core/felt"
)

type Program struct {
	// Prime is fixed to be 0x800000000000011000000000000000000000000000000000000000000000001 and wont fit in a f.Felt
	Bytecode        []f.Felt `json:"bytecode"`
	CompilerVersion string   `json:"compiler_version"`
	// todo(rodro): Add remaining Json fields
	// Hints
	// EntryPointsByType
	// Contructors
	// L1 Headers
}

func ProgramFromFile(pathToFile string) (*Program, error) {
	content, error := os.ReadFile(pathToFile)
	if error != nil {
		return nil, error
	}
	return ProgramFromJSON(content)
}

func ProgramFromJSON(content json.RawMessage) (*Program, error) {
	var program Program
	return &program, json.Unmarshal(content, &program)
}
