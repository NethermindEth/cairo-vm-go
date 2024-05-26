package utils

var IV = [8]uint32{
	0x6A09E667,
	0xBB67AE85,
	0x3C6EF372,
	0xA54FF53A,
	0x510E527F,
	0x9B05688C,
	0x1F83D9AB,
	0x5BE0CD19,
}

var SIGMA = [10][16]uint32{
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	{14, 10, 4, 8, 9, 15, 13, 6, 1, 12, 0, 2, 11, 7, 5, 3},
	{11, 8, 12, 0, 5, 2, 15, 13, 10, 14, 3, 6, 7, 1, 9, 4},
	{7, 9, 3, 1, 13, 12, 11, 14, 2, 6, 5, 10, 4, 0, 15, 8},
	{9, 0, 5, 7, 2, 4, 10, 15, 14, 1, 11, 12, 6, 8, 3, 13},
	{2, 12, 6, 10, 0, 11, 8, 3, 4, 13, 7, 5, 15, 14, 1, 9},
	{12, 5, 1, 15, 14, 13, 4, 10, 0, 7, 6, 3, 9, 2, 8, 11},
	{13, 11, 7, 14, 12, 1, 3, 9, 5, 0, 15, 4, 8, 6, 2, 10},
	{6, 15, 14, 9, 11, 3, 0, 8, 12, 2, 13, 7, 1, 4, 10, 5},
	{10, 2, 8, 4, 7, 6, 1, 5, 15, 11, 9, 14, 3, 12, 13, 0},
}

func rightRot(value, n uint32) uint32 {
	return (value >> n) | ((value & ((1 << n) - 1)) << (32 - n))
}

func mix(a, b, c, d, m0, m1 uint32) (uint32, uint32, uint32, uint32) {
	a = a + b + m0
	d = rightRot(d^a, 16)
	c = c + d
	b = rightRot(b^c, 12)
	a = a + b + m1
	d = rightRot(d^a, 8)
	c = c + d
	b = rightRot(b^c, 7)

	return a, b, c, d
}

func blakeRound(state, message []uint32, sigma [16]uint32) []uint32 {
	state[0], state[4], state[8], state[12] = mix(
		state[0], state[4], state[8], state[12], message[sigma[0]], message[sigma[1]],
	)
	state[1], state[5], state[9], state[13] = mix(
		state[1], state[5], state[9], state[13], message[sigma[2]], message[sigma[3]],
	)
	state[2], state[6], state[10], state[14] = mix(
		state[2], state[6], state[10], state[14], message[sigma[4]], message[sigma[5]],
	)
	state[3], state[7], state[11], state[15] = mix(
		state[3], state[7], state[11], state[15], message[sigma[6]], message[sigma[7]],
	)

	state[0], state[5], state[10], state[15] = mix(
		state[0], state[5], state[10], state[15], message[sigma[8]], message[sigma[9]],
	)
	state[1], state[6], state[11], state[12] = mix(
		state[1], state[6], state[11], state[12], message[sigma[10]], message[sigma[11]],
	)
	state[2], state[7], state[8], state[13] = mix(
		state[2], state[7], state[8], state[13], message[sigma[12]], message[sigma[13]],
	)
	state[3], state[4], state[9], state[14] = mix(
		state[3], state[4], state[9], state[14], message[sigma[14]], message[sigma[15]],
	)

	return state
}

func Blake2sCompress(
	h, message []uint32, t0, t1, f0, f1 uint32,
) [8]uint32 {
	state := append(h, IV[:4]...)
	state = append(state, IV[4]^t0, IV[5]^t1, IV[6]^f0, IV[7]^f1)

	for i := 0; i < 10; i++ {
		state = blakeRound(state, message, SIGMA[i])
	}

	var result [8]uint32
	for i, v := range h {
		result[i] = v ^ state[i] ^ state[i+8]
	}

	return result
}
