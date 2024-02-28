package main

import (
	"errors"
	"fmt"
	"strings"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"github.com/urfave/cli/v2"
)

// instFieldsProgram implements an "inst-fields" subcommand.
type instFieldsProgram struct{}

func (p *instFieldsProgram) Action(ctx *cli.Context) error {
	s := ctx.Args().Get(0)
	if s == "" {
		return errors.New("expected 1 non-empty positional argument")
	}

	felt, err := new(f.Element).SetString(s)
	if err != nil {
		return fmt.Errorf("parsing %q argument: %w", s, err)
	}
	if !felt.IsUint64() {
		return errors.New("instruction bytes overflow uint64")
	}

	u64 := felt.Uint64()

	fmt.Printf("uint64 value: %v\n", u64)

	// We don't use the assembler's package code here to make it possible
	// to use this dumper tool even if assembler package can't validate
	// the input. Unlike the assembler package, this tool doesn't care
	// if the provided bits are valid or not.
	// It will split them into "fields" expected by the CASM instruction encoding.

	type instField struct {
		name   string
		width  int // in bits
		signed bool
	}
	encodingList := []instField{
		{"off_dst", 16, true},
		{"off_op0", 16, true},
		{"off_op1", 16, true},
		{"dst_reg", 1, false},
		{"op0_reg", 1, false},
		{"op1_src", 3, false},
		{"res_logic", 2, false},
		{"pc_update", 3, false},
		{"ap_update", 2, false},
		{"opcode", 3, false},
	}

	const onesMask = ^uint64(0)

	var chunks []string

	offset := int(0)
	for _, field := range encodingList {
		mask := onesMask >> (64 - field.width)
		fieldBits := (u64 >> offset) & mask
		if field.signed {
			fmt.Printf("%s: %v (%b)\n", field.name, int16(fieldBits), fieldBits)
		} else {
			fmt.Printf("%s: %v (%b)\n", field.name, fieldBits, fieldBits)
		}
		chunks = append(chunks, fmt.Sprintf("%b", fieldBits))
		offset += field.width
	}

	fmt.Printf("bits: %s\n", strings.Join(chunks, " "))

	return nil
}
