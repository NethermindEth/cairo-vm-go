package vm

import (
	"fmt"

	f "github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// Represents a write-once Memory Cell
type Cell struct {
	Value    f.Element
	Accessed bool
}

func (cell *Cell) Read() f.Element {
	cell.Accessed = true
	return cell.Value
}

func (cell *Cell) Write(value f.Element) error {
	if cell.Accessed {
		return fmt.Errorf("rewriting cell, old value: %d new value: %d", cell.Value, value)
	}
	cell.Accessed = true
	cell.Value = value
	return nil
}

// Represents the whole VM memory divided by segments
type Memory struct {
}
