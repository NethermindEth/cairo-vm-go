package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/NethermindEth/cairo-vm-go/pkg/disasm"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"

	"github.com/urfave/cli/v2"
)

// disasmProgram implements a "disasm" subcommand.
type disasmProgram struct {
	pathToFile  string
	bytecodeKey string

	rawCasm map[string]any

	bytecode []*fp.Element

	disassembled *disasm.Program
}

func (p *disasmProgram) Action(ctx *cli.Context) error {
	p.pathToFile = ctx.Args().Get(0)
	if p.pathToFile == "" {
		return fmt.Errorf("path to casm file not set")
	}

	type step struct {
		name string
		fn   func() error
	}
	steps := []step{
		{"unmarshal casm file", p.unmarshalCasmFileStep},
		{"load bytecode", p.loadBytecodeStep},
		{"disassemble", p.disassembleStep},
		{"print", p.printStep},
	}
	for _, s := range steps {
		if err := s.fn(); err != nil {
			return fmt.Errorf("%s: %w", s.name, err)
		}
	}

	return nil
}

func (p *disasmProgram) unmarshalCasmFileStep() error {
	data, err := os.ReadFile(p.pathToFile)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &p.rawCasm); err != nil {
		return err
	}
	return nil
}

func (p *disasmProgram) loadBytecodeStep() error {
	// Since different versions of CASM files may store bytecode at different places
	// (e.g. "data" in Cairo0 and "bytecode" in Cairo1),
	// we allow the user to specify the bytecode array location.
	// By default, this value will be equal to the default supported version location
	// (Cairo0 for now and Cairo1 in the future).
	keys := strings.Split(p.bytecodeKey, ".")

	v := lookupKeys(p.rawCasm, keys...)
	if v == nil {
		return fmt.Errorf("key %q doesn't lead to a bytecode", p.bytecodeKey)
	}

	slice, ok := v.([]any)
	if !ok {
		return fmt.Errorf("%q: expected a slice of strings", p.bytecodeKey)
	}

	p.bytecode = make([]*fp.Element, 0, len(slice))
	for i, s := range slice {
		s, ok := s.(string)
		if !ok {
			return fmt.Errorf("%q: expected a slice of strings, found %T", p.bytecodeKey, slice[i])
		}
		felt, err := new(fp.Element).SetString(s)
		if err != nil {
			return fmt.Errorf("%q[%d]: parse %q: %w", p.bytecodeKey, i, s, err)
		}
		p.bytecode = append(p.bytecode, felt)
	}

	return nil
}

func (p *disasmProgram) disassembleStep() error {
	prog, err := disasm.FromBytecode(disasm.Config{
		Bytecode: p.bytecode,
		Indent:   4,
	})
	if err != nil {
		return err
	}
	p.disassembled = prog
	return nil
}

func (p *disasmProgram) printStep() error {
	for _, l := range p.disassembled.Lines {
		if len(l.Text) == 0 {
			fmt.Printf("// %s\n", strings.Join(l.Comments, "; "))
			continue
		}
		if len(l.Comments) == 0 {
			fmt.Printf("%s;\n", l.Text)
		} else {
			fmt.Printf("%s; // %s\n", l.Text, strings.Join(l.Comments, "; "))
		}
	}
	return nil
}
