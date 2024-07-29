package utils

import "fmt"

func ComputeMessageSchedule(input []uint32) ([]uint32, error) {
	// def compute_message_schedule(message: List[int]) -> List[int]:
	// w = list(message)
	// assert len(w) == 16

	// for i in range(16, 64):
	//     s0 = right_rot(w[i - 15], 7) ^ right_rot(w[i - 15], 18) ^ (w[i - 15] >> 3)
	//     s1 = right_rot(w[i - 2], 17) ^ right_rot(w[i - 2], 19) ^ (w[i - 2] >> 10)
	//     w.append((w[i - 16] + s0 + w[i - 7] + s1) % 2**32)

	// return w
	if len(input) != 16 {
		return nil, fmt.Errorf("input length must be 16, got %d", len(input))
	}

	fmt.Println((input))

	w := make([]uint32, 64)
	copy(w, input)

	for i := 16; i < 64; i++ {
		s0 := RightRot(w[i-15], 7) ^ RightRot(w[i-15], 18) ^ (w[i-15] >> 3)
		s1 := RightRot(w[i-2], 17) ^ RightRot(w[i-2], 19) ^ (w[i-2] >> 10)
		w[i] = (w[i-16] + s0 + w[i-7] + s1)
	}

	return w, nil
}

func Sha256Compress(state [8]uint32, w []uint32) []uint32 {
	// def sha2_compress_function(state: List[int], w: List[int]) -> List[int]:
	//     a, b, c, d, e, f, g, h = state

	//     for i in range(64):
	//         s0 = right_rot(a, 2) ^ right_rot(a, 13) ^ right_rot(a, 22)
	//         s1 = right_rot(e, 6) ^ right_rot(e, 11) ^ right_rot(e, 25)
	//         ch = (e & f) ^ ((~e) & g)
	//         temp1 = (h + s1 + ch + ROUND_CONSTANTS[i] + w[i]) % 2**32
	//         maj = (a & b) ^ (a & c) ^ (b & c)
	//         temp2 = (s0 + maj) % 2**32

	//         h = g
	//         g = f
	//         f = e
	//         e = (d + temp1) % 2**32
	//         d = c
	//         c = b
	//         b = a
	//         a = (temp1 + temp2) % 2**32

	// # Add the compression result to the original state.
	// return [
	//
	//	(state[0] + a) % 2**32,
	//	(state[1] + b) % 2**32,
	//	(state[2] + c) % 2**32,
	//	(state[3] + d) % 2**32,
	//	(state[4] + e) % 2**32,
	//	(state[5] + f) % 2**32,
	//	(state[6] + g) % 2**32,
	//	(state[7] + h) % 2**32,
	//
	// ]
	k := []uint32{
		0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
		0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
		0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
		0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
		0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
		0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
		0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
		0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
	}

	a, b, c, d, e, f, g, h := state[0], state[1], state[2], state[3], state[4], state[5], state[6], state[7]

	for i := 0; i < 64; i++ {
		S1 := RightRot(e, 6) ^ RightRot(e, 11) ^ RightRot(e, 25)
		ch := (e & f) ^ ((^e) & g)
		temp1 := h + S1 + ch + k[i] + w[i]
		S0 := RightRot(a, 2) ^ RightRot(a, 13) ^ RightRot(a, 22)
		maj := (a & b) ^ (a & c) ^ (b & c)
		temp2 := S0 + maj

		h = g
		g = f
		f = e
		e = d + temp1
		d = c
		c = b
		b = a
		a = temp1 + temp2
	}

	return []uint32{
		state[0] + a, state[1] + b, state[2] + c, state[3] + d,
		state[4] + e, state[5] + f, state[6] + g, state[7] + h,
	}
}
