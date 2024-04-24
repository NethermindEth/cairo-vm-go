package zero

const (
	// ------ Math hints related code ------
	// This is a block for hint code strings where there is a single
	// hint per function it belongs to (with some exceptions like testAssignCode).
	isLeFeltCode       string = "memory[ap] = 0 if (ids.a % PRIME) <= (ids.b % PRIME) else 1"
	assertLtFeltCode   string = "from starkware.cairo.common.math_utils import assert_integer\nassert_integer(ids.a)\nassert_integer(ids.b)\nassert (ids.a % PRIME) < (ids.b % PRIME), \\\n    f'a = {ids.a % PRIME} is not less than b = {ids.b % PRIME}.'"
	assertNotZeroCode  string = "from starkware.cairo.common.math_utils import assert_integer\nassert_integer(ids.value)\nassert ids.value % PRIME != 0, f'assert_not_zero failed: {ids.value} = 0.'"
	assertNNCode       string = "from starkware.cairo.common.math_utils import assert_integer\nassert_integer(ids.a)\nassert 0 <= ids.a % PRIME < range_check_builtin.bound, f'a = {ids.a} is out of range.'"
	assertNotEqualCode string = "from starkware.cairo.lang.vm.relocatable import RelocatableValue\nboth_ints = isinstance(ids.a, int) and isinstance(ids.b, int)\nboth_relocatable = (\n    isinstance(ids.a, RelocatableValue) and isinstance(ids.b, RelocatableValue) and\n    ids.a.segment_index == ids.b.segment_index)\nassert both_ints or both_relocatable, \\\n    f'assert_not_equal failed: non-comparable values: {ids.a}, {ids.b}.'\nassert (ids.a - ids.b) % PRIME != 0, f'assert_not_equal failed: {ids.a} = {ids.b}.'"
	assert250bits      string = "from starkware.cairo.common.math_utils import as_int\n\n# Correctness check.\nvalue = as_int(ids.value, PRIME) % PRIME\nassert value < ids.UPPER_BOUND, f'{value} is outside of the range [0, 2**250).'\n\n# Calculation for the assertion.\nids.high, ids.low = divmod(ids.value, ids.SHIFT)"

	// This is a very simple Cairo0 hint that allows us to test
	// the identifier resolution code.
	// Depending on the context, ids.a may be a complex reference.
	testAssignCode string = "memory[ap] = ids.a"

	// assert_le_felt() hints.
	assertLeFeltCode          string = "import itertools\n\nfrom starkware.cairo.common.math_utils import assert_integer\nassert_integer(ids.a)\nassert_integer(ids.b)\na = ids.a % PRIME\nb = ids.b % PRIME\nassert a <= b, f'a = {a} is not less than or equal to b = {b}.'\n\n# Find an arc less than PRIME / 3, and another less than PRIME / 2.\nlengths_and_indices = [(a, 0), (b - a, 1), (PRIME - 1 - b, 2)]\nlengths_and_indices.sort()\nassert lengths_and_indices[0][0] <= PRIME // 3 and lengths_and_indices[1][0] <= PRIME // 2\nexcluded = lengths_and_indices[2][1]\n\nmemory[ids.range_check_ptr + 1], memory[ids.range_check_ptr + 0] = (\n    divmod(lengths_and_indices[0][0], ids.PRIME_OVER_3_HIGH))\nmemory[ids.range_check_ptr + 3], memory[ids.range_check_ptr + 2] = (\n    divmod(lengths_and_indices[1][0], ids.PRIME_OVER_2_HIGH))"
	assertLeFeltExcluded0Code string = "memory[ap] = 1 if excluded != 0 else 0"
	assertLeFeltExcluded1Code string = "memory[ap] = 1 if excluded != 1 else 0"
	assertLeFeltExcluded2Code string = "assert excluded == 2"

	// is_nn() hints.
	isNNCode           string = "memory[ap] = 0 if 0 <= (ids.a % PRIME) < range_check_builtin.bound else 1"
	isNNOutOfRangeCode string = "memory[ap] = 0 if 0 <= ((-ids.a - 1) % PRIME) < range_check_builtin.bound else 1"

	// This is a rare case when some hint is used in more than one place.
	// isPositive is used in sign() and abs_value() functions.
	isPositiveCode string = "from starkware.cairo.common.math_utils import is_positive\nids.is_positive = 1 if is_positive(\n    value=ids.value, prime=PRIME, rc_bound=range_check_builtin.bound) else 0"

	// split_int() hints.
	splitIntAssertRange string = "assert ids.value == 0, 'split_int(): value is out of range.'"
	splitIntCode        string = "memory[ids.output] = res = (int(ids.value) % PRIME) % ids.base\nassert res < ids.bound, f'split_int(): Limb {res} is out of range.'"
	signedDivRemCode    string = "from starkware.cairo.common.math_utils import as_int, assert_integer\nassert_integer(ids.div)\nassert 0 < ids.div <= PRIME // range_check_builtin.bound, f'div={hex(ids.div)} is out of the valid range.'\nassert_integer(ids.bound)\nassert ids.bound <= range_check_builtin.bound // 2, f'bound={hex(ids.bound)} is out of the valid range.'\nint_value = as_int(ids.value, PRIME)\nq, ids.r = divmod(int_value, ids.div)\nassert -ids.bound <= q < ids.bound, f'{int_value} / {ids.div} = {q} is out of the range [{-ids.bound}, {ids.bound}).'\nids.biased_q = q + ids.bound"

	// pow hints
	powCode string = "ids.locs.bit = (ids.prev_locs.exp % PRIME) & 1"

	unsignedDivRemCode string = "from starkware.cairo.common.math_utils import assert_integer\nassert_integer(ids.div)\nassert 0 < ids.div <= PRIME // range_check_builtin.bound, \\\n    f'div={hex(ids.div)} is out of the valid range.'\nids.q, ids.r = divmod(ids.value, ids.div)"

	// split_felt() hints.
	splitFeltCode string = "from starkware.cairo.common.math_utils import assert_integer\nassert ids.MAX_HIGH < 2**128 and ids.MAX_LOW < 2**128\nassert PRIME - 1 == ids.MAX_HIGH * 2**128 + ids.MAX_LOW\nassert_integer(ids.value)\nids.low = ids.value & ((1 << 128) - 1)\nids.high = ids.value >> 128"

	// sqrt() hint
	sqrtCode string = "from starkware.python.math_utils import isqrt\nvalue = ids.value % PRIME\nassert value < 2 ** 250, f\"value={value} is outside of the range [0, 2**250).\"\nassert 2 ** 250 < PRIME\nids.root = isqrt(value)"

	// ------ Uint256 hints related code ------
	uint256AddCode            string = "sum_low = ids.a.low + ids.b.low\nids.carry_low = 1 if sum_low >= ids.SHIFT else 0\nsum_high = ids.a.high + ids.b.high + ids.carry_low\nids.carry_high = 1 if sum_high >= ids.SHIFT else 0"
	uint256AddLowCode         string = "sum_low = ids.a.low + ids.b.low\nids.carry_low = 1 if sum_low >= ids.SHIFT else 0"
	split64Code               string = "ids.low = ids.a & ((1<<64) - 1)\nids.high = ids.a >> 64"
	uint256SignedNNCode       string = "memory[ap] = 1 if 0 <= (ids.a.high % PRIME) < 2 ** 127 else 0"
	uint256UnsignedDivRemCode string = "a = (ids.a.high << 128) + ids.a.low\ndiv = (ids.div.high << 128) + ids.div.low\nquotient, remainder = divmod(a, div)\nids.quotient.low = quotient & ((1 << 128) - 1)\nids.quotient.high = quotient >> 128\nids.remainder.low = remainder & ((1 << 128) - 1)\nids.remainder.high = remainder >> 128"
	uint256SqrtCode           string = "from starkware.python.math_utils import isqrt\nn = (ids.n.high << 128) + ids.n.low\nroot = isqrt(n)\nassert 0 <= root < 2 ** 128\nids.root.low = root\nids.root.high = 0"
	uint256MulDivModCode      string = "a = (ids.a.high << 128) + ids.a.low/n b = (ids.b.high << 128) + ids.b.low/n div = (ids.div.high << 128) + ids.div.low/n quotient, remainder = divmod(a * b, div)/n ids.quotient_low.low = quotient & ((1 << 128) - 1)/n ids.quotient_low.high = (quotient >> 128) & ((1 << 128) - 1)/n ids.quotient_high.low = (quotient >> 256) & ((1 << 128) - 1)/n ids.quotient_high.high = quotient >> 384/n ids.remainder.low = remainder & ((1 << 128) - 1)/n ids.remainder.high = remainder >> 128"

	// ------ Usort hints related code ------
	usortEnterScopeCode               string = "vm_enter_scope(dict(__usort_max_size = globals().get('__usort_max_size')))"
	usortVerifyMultiplicityAssertCode string = "assert len(positions) == 0"
	usortVerifyCode                   string = "last_pos = 0\npositions = positions_dict[ids.value][::-1]"
	usortVerifyMultiplicityBodyCode   string = "current_pos = positions.pop()\nids.next_item_index = current_pos - last_pos\nlast_pos = current_pos + 1"

	// ------ Elliptic Curve hints related code ------
	ecNegateCode            string = "from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack\n\ny = pack(ids.point.y, PRIME) % SECP_P\n# The modulo operation in python always returns a nonnegative number.\nvalue = (-y) % SECP_P"
	nondetBigint3V1Code     string = "from starkware.cairo.common.cairo_secp.secp_utils import split\n\nsegments.write_arg(ids.res.address_, split(value))"
	fastEcAddAssignNewYCode string = "value = new_y = (slope * (x0 - new_x) - y0) % SECP_P"
	fastEcAddAssignNewXCode string = "from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack\n\nslope = pack(ids.slope, PRIME)\nx0 = pack(ids.point0.x, PRIME)\nx1 = pack(ids.point1.x, PRIME)\ny0 = pack(ids.point0.y, PRIME)\n\nvalue = new_x = (pow(slope, 2, SECP_P) - x0 - x1) % SECP_P"
	ecDoubleSlopeV1Code     string = "from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack\nfrom starkware.python.math_utils import ec_double_slope\n\n# Compute the slope.\nx = pack(ids.point.x, PRIME)\ny = pack(ids.point.y, PRIME)\nvalue = slope = ec_double_slope(point=(x, y), alpha=0, p=SECP_P)"
	ecDoubleAssignNewXV1    string = "from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack\n\nslope = pack(ids.slope, PRIME)\nx = pack(ids.point.x, PRIME)\ny = pack(ids.point.y, PRIME)\n\nvalue = new_x = (pow(slope, 2, SECP_P) - 2 * x) % SECP_P"

	// ------ Signature hints related code ------
	verifyECDSASignatureCode  string = "ecdsa_builtin.add_signature(ids.ecdsa_ptr.address_, (ids.signature_r, ids.signature_s))"
	getPointFromXCode         string = "from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack\nx_cube_int = pack(ids.x_cube, PRIME) % SECP_P\ny_square_int = (x_cube_int + ids.BETA) % SECP_P\ny = pow(y_square_int, (SECP_P + 1) // 4, SECP_P)\nif ids.v % 2 == y % 2:\nvalue = y\nelse:\nvalue = (-y) % SECP_P"
	divModNSafeDivCode        string = "value = k = safe_div(res * b - a, N)"
	importSecp256R1PCode      string = "from starkware.cairo.common.cairo_secp.secp256r1_utils import SECP256R1_P as SECP_P"
	verifyZeroCode            string = "from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack\n\nq, r = divmod(pack(ids.val, PRIME), SECP_P)\nassert r == 0, f\"verify_zero: Invalid input {ids.val.d0, ids.val.d1, ids.val.d2}.\"\nids.q = q % PRIME"
	divModNPackedDivmodV1Code string = "from starkware.cairo.common.cairo_secp.secp_utils import N, pack\nfrom starkware.python.math_utils import div_mod, safe_div\n\na = pack(ids.a, PRIME)\nb = pack(ids.b, PRIME)\nvalue = res = div_mod(a, b, N)"

	// ------ Blake Hash hints related code ------
	blake2sAddUint256BigendCode string = "B = 32\nMASK = 2 ** 32 - 1\nsegments.write_arg(ids.data, [(ids.high >> (B * (3 - i))) & MASK for i in range(4)])\nsegments.write_arg(ids.data + 4, [(ids.low >> (B * (3 - i))) & MASK for i in range(4)])"
	blake2sAddUint256Code       string = "B = 32\nMASK = 2 ** 32 - 1\nsegments.write_arg(ids.data, [(ids.low >> (B * i)) & MASK for i in range(4)])\nsegments.write_arg(ids.data + 4, [(ids.high >> (B * i)) & MASK for i in range(4)])"

	// ------ Keccak hints related code ------

	keccakWriteArgs string = "segments.write_arg(ids.inputs, [ids.low % 2 ** 64, ids.low // 2 ** 64])\nsegments.write_arg(ids.inputs + 2, [ids.high % 2 ** 64, ids.high // 2 ** 64])"

	// ------ Dictionaries hints related code ------
	squashDictInnerAssertLenKeys string = "assert len(keys) == 0"

	// ------ Other hints related code ------
	allocSegmentCode     string = "memory[ap] = segments.add()"
	memcpyEnterScopeCode string = "vm_enter_scope({'n': ids.len})"
	vmEnterScopeCode     string = "vm_enter_scope()"
	vmExitScopeCode      string = "vm_exit_scope()"
)
