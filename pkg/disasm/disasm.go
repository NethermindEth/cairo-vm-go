package disasm

import (
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

type Program struct {
	Lines []Line
}

type Line struct {
	Text     string
	Comments []string
}

type Config struct {
	Bytecode []*fp.Element

	Indent int
}

func FromBytecode(config Config) (*Program, error) {
	d := &disassembler{
		bytecode: config.Bytecode,
		config:   config,
	}
	return d.Disassemble()
}
