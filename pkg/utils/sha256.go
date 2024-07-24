package utils

import (
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func ComputeMessageSchedule(w []fp.Element) []fp.Element {
	// def compute_message_schedule(message: List[int]) -> List[int]:
	// w = list(message)
	// assert len(w) == 16

	// for i in range(16, 64):
	//     s0 = right_rot(w[i - 15], 7) ^ right_rot(w[i - 15], 18) ^ (w[i - 15] >> 3)
	//     s1 = right_rot(w[i - 2], 17) ^ right_rot(w[i - 2], 19) ^ (w[i - 2] >> 10)
	//     w.append((w[i - 16] + s0 + w[i - 7] + s1) % 2**32)

	// return w
	var emptyArray []fp.Element
	return emptyArray
}

func Sha256Compress(IV []uint32, w []fp.Element) []fp.Element {
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
	var emptyArray []fp.Element
	return emptyArray
}
