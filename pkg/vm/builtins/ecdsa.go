package builtins

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	starkcurve "github.com/consensys/gnark-crypto/ecc/stark-curve"
	ecdsa "github.com/consensys/gnark-crypto/ecc/stark-curve/ecdsa"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

const ECDSAName = "ecdsa"
const cellsPerECDSA = 2
const inputCellsPerECDSA = 2 //Public key and msg

type ECDSA struct {
	signatures map[uint64]ecdsa.Signature
}

func (e *ECDSA) InferValue(segment *memory.Segment, offset uint64) error {
	ecdsaIndex := offset % cellsPerECDSA
	// input cell
	pubOffset := offset - ecdsaIndex
	msg_offset := pubOffset + 1

	pub := segment.Peek(offset)
	if !pub.Known() {
		return fmt.Errorf("cannot infer value: input value at offset %d is unknown", pubOffset)
	}

	pubX, err := pub.FieldElement() //X element of the sig
	if err != nil {
		return err
	}

	msg := segment.Peek(offset)
	if !msg.Known() {
		return fmt.Errorf("cannot infer value: input value at offset %d is unknown", msg_offset)
	}
	msgField, err := msg.FieldElement()
	if err != nil {
		return err
	}

	//Sig verification
	pubY, err := recoverY(pubX)
	if err != nil {
		return err
	}

	key := starkcurve.G1Affine{X: *pubX, Y: *pubY}
	if !key.IsOnCurve() {
		return fmt.Errorf("Key is not on curve")
	}

	pubKey := &ecdsa.PublicKey{A: key}
	sig, ok := e.signatures[pubOffset]
	if !ok {
		return fmt.Errorf("Signature is missing form ECDA builtin")
	}
	msgBytes := msgField.Bytes()
	valid, err := pubKey.Verify(sig.Bytes(), msgBytes[:], nil)
	if err != nil {
		return err
	}

	if !valid {
		return fmt.Errorf("Signature is not valid")
	}

	//TODO: Get r, s, pub and hash

	return nil
}

// "code": "ecdsa_builtin.add_signature(ids.ecdsa_ptr.address_, (ids.signature_r, ids.signature_s))",
func (e *ECDSA) addSignature(pubOffset uint64, r, s fp.Element) error {
	bytes := make([]byte, 0, 64)
	rBytes := r.Bytes()
	bytes = append(bytes, rBytes[:]...)
	sBytes := s.Bytes()
	bytes = append(bytes, sBytes[:]...)

	sig := ecdsa.Signature{}
	_, err := sig.SetBytes(bytes)
	if err != nil {
		return err
	}

	e.signatures[pubOffset] = sig
	return nil
}

func (e *ECDSA) CheckWrite(segment *memory.Segment, offset uint64, value *memory.MemoryValue) error {
	return nil
}

func (e *ECDSA) String() string {
	return ECDSAName
}

// recoverY recovers the y coordinate of x. True y can be either y or -y
func recoverY(x *fp.Element) (*fp.Element, error) {
	ALPHA := fp.NewElement(1)
	BETA := fp.Element{}
	_, _ = BETA.SetString("3141592653589793238462643383279502884197169399375105820974944592307816406665")
	// y_squared = (x * x * x + ALPHA * x + BETA) % FIELD_PRIME
	x2 := new(fp.Element).Mul(x, x)
	x3 := x2.Mul(x2, x)
	a := new(fp.Element).Mul(&ALPHA, x)
	x3.Add(x3, a)
	x3.Add(x3, &BETA)
	y := x3.Sqrt(x3)
	if y == nil {
		return nil, fmt.Errorf("Invalid Public key")
	}
	return y, nil
}
