package zero

const (
	// This is a very simple Cairo0 hint that allows us to test the identifier resolution code.
	// Depending on the context, ids.a may be a complex reference.
	testAssignCode string = "memory[ap] = ids.a"

	// ------ Math hints related code ------
	// This is a block for hint code strings where there is a single hint per function it belongs to.
	isLeFeltCode       string = "memory[ap] = 0 if (ids.a % PRIME) <= (ids.b % PRIME) else 1"
	assertLtFeltCode   string = "from starkware.cairo.common.math_utils import assert_integer\nassert_integer(ids.a)\nassert_integer(ids.b)\nassert (ids.a % PRIME) < (ids.b % PRIME), \\\n    f'a = {ids.a % PRIME} is not less than b = {ids.b % PRIME}.'"
	assertNotZeroCode  string = "from starkware.cairo.common.math_utils import assert_integer\nassert_integer(ids.value)\nassert ids.value % PRIME != 0, f'assert_not_zero failed: {ids.value} = 0.'"
	assertNNCode       string = "from starkware.cairo.common.math_utils import assert_integer\nassert_integer(ids.a)\nassert 0 <= ids.a % PRIME < range_check_builtin.bound, f'a = {ids.a} is out of range.'"
	assertNotEqualCode string = "from starkware.cairo.lang.vm.relocatable import RelocatableValue\nboth_ints = isinstance(ids.a, int) and isinstance(ids.b, int)\nboth_relocatable = (\n    isinstance(ids.a, RelocatableValue) and isinstance(ids.b, RelocatableValue) and\n    ids.a.segment_index == ids.b.segment_index)\nassert both_ints or both_relocatable, \\\n    f'assert_not_equal failed: non-comparable values: {ids.a}, {ids.b}.'\nassert (ids.a - ids.b) % PRIME != 0, f'assert_not_equal failed: {ids.a} = {ids.b}.'"
	assert250bitsCode  string = "from starkware.cairo.common.math_utils import as_int\n\n# Correctness check.\nvalue = as_int(ids.value, PRIME) % PRIME\nassert value < ids.UPPER_BOUND, f'{value} is outside of the range [0, 2**250).'\n\n# Calculation for the assertion.\nids.high, ids.low = divmod(ids.value, ids.SHIFT)"

	// assert_le_felt() hints
	assertLeFeltCode          string = "import itertools\n\nfrom starkware.cairo.common.math_utils import assert_integer\nassert_integer(ids.a)\nassert_integer(ids.b)\na = ids.a % PRIME\nb = ids.b % PRIME\nassert a <= b, f'a = {a} is not less than or equal to b = {b}.'\n\n# Find an arc less than PRIME / 3, and another less than PRIME / 2.\nlengths_and_indices = [(a, 0), (b - a, 1), (PRIME - 1 - b, 2)]\nlengths_and_indices.sort()\nassert lengths_and_indices[0][0] <= PRIME // 3 and lengths_and_indices[1][0] <= PRIME // 2\nexcluded = lengths_and_indices[2][1]\n\nmemory[ids.range_check_ptr + 1], memory[ids.range_check_ptr + 0] = (\n    divmod(lengths_and_indices[0][0], ids.PRIME_OVER_3_HIGH))\nmemory[ids.range_check_ptr + 3], memory[ids.range_check_ptr + 2] = (\n    divmod(lengths_and_indices[1][0], ids.PRIME_OVER_2_HIGH))"
	assertLeFeltExcluded0Code string = "memory[ap] = 1 if excluded != 0 else 0"
	assertLeFeltExcluded1Code string = "memory[ap] = 1 if excluded != 1 else 0"
	assertLeFeltExcluded2Code string = "assert excluded == 2"

	// is_nn() hints
	isNNCode           string = "memory[ap] = 0 if 0 <= (ids.a % PRIME) < range_check_builtin.bound else 1"
	isNNOutOfRangeCode string = "memory[ap] = 0 if 0 <= ((-ids.a - 1) % PRIME) < range_check_builtin.bound else 1"

	// This is a rare case when some hint is used in more than one place.
	// isPositive is used in sign() and abs_value() functions.
	isPositiveCode string = "from starkware.cairo.common.math_utils import is_positive\nids.is_positive = 1 if is_positive(\n    value=ids.value, prime=PRIME, rc_bound=range_check_builtin.bound) else 0"

	// split_int() hints
	splitIntAssertRangeCode string = "assert ids.value == 0, 'split_int(): value is out of range.'"
	splitIntCode            string = "memory[ids.output] = res = (int(ids.value) % PRIME) % ids.base\nassert res < ids.bound, f'split_int(): Limb {res} is out of range.'"

	// pow() hint
	powCode       string = "ids.locs.bit = (ids.prev_locs.exp % PRIME) & 1"
	signedPowCode string = "assert ids.base != 0, 'Cannot raise 0 to a negative power.'"

	// div_rem() hints
	signedDivRemCode   string = "from starkware.cairo.common.math_utils import as_int, assert_integer\n\nassert_integer(ids.div)\nassert 0 < ids.div <= PRIME // range_check_builtin.bound, \\\n    f'div={hex(ids.div)} is out of the valid range.'\n\nassert_integer(ids.bound)\nassert ids.bound <= range_check_builtin.bound // 2, \\\n    f'bound={hex(ids.bound)} is out of the valid range.'\n\nint_value = as_int(ids.value, PRIME)\nq, ids.r = divmod(int_value, ids.div)\n\nassert -ids.bound <= q < ids.bound, \\\n    f'{int_value} / {ids.div} = {q} is out of the range [{-ids.bound}, {ids.bound}).'\n\nids.biased_q = q + ids.bound"
	unsignedDivRemCode string = "from starkware.cairo.common.math_utils import assert_integer\nassert_integer(ids.div)\nassert 0 < ids.div <= PRIME // range_check_builtin.bound, \\\n    f'div={hex(ids.div)} is out of the valid range.'\nids.q, ids.r = divmod(ids.value, ids.div)"

	// split_felt() hint
	splitFeltCode string = "from starkware.cairo.common.math_utils import assert_integer\nassert ids.MAX_HIGH < 2**128 and ids.MAX_LOW < 2**128\nassert PRIME - 1 == ids.MAX_HIGH * 2**128 + ids.MAX_LOW\nassert_integer(ids.value)\nids.low = ids.value & ((1 << 128) - 1)\nids.high = ids.value >> 128"

	// sqrt() hint
	sqrtCode string = "from starkware.python.math_utils import isqrt\nvalue = ids.value % PRIME\nassert value < 2 ** 250, f\"value={value} is outside of the range [0, 2**250).\"\nassert 2 ** 250 < PRIME\nids.root = isqrt(value)"

	// is_quad_residue() hint
	isQuadResidueCode string = "from starkware.crypto.signature.signature import FIELD_PRIME\nfrom starkware.python.math_utils import div_mod, is_quad_residue, sqrt\n\nx = ids.x\nif is_quad_residue(x, FIELD_PRIME):\n    ids.y = sqrt(x, FIELD_PRIME)\nelse:\n    ids.y = sqrt(div_mod(x, 3, FIELD_PRIME), FIELD_PRIME)"

	// ------ Uint256 hints related code ------
	uint256AddCode            string = "sum_low = ids.a.low + ids.b.low\nids.carry_low = 1 if sum_low >= ids.SHIFT else 0\nsum_high = ids.a.high + ids.b.high + ids.carry_low\nids.carry_high = 1 if sum_high >= ids.SHIFT else 0"
	split64Code               string = "ids.low = ids.a & ((1<<64) - 1)\nids.high = ids.a >> 64"
	uint256SignedNNCode       string = "memory[ap] = 1 if 0 <= (ids.a.high % PRIME) < 2 ** 127 else 0"
	uint256UnsignedDivRemCode string = "a = (ids.a.high << 128) + ids.a.low\ndiv = (ids.div.high << 128) + ids.div.low\nquotient, remainder = divmod(a, div)\n\nids.quotient.low = quotient & ((1 << 128) - 1)\nids.quotient.high = quotient >> 128\nids.remainder.low = remainder & ((1 << 128) - 1)\nids.remainder.high = remainder >> 128"
	uint256SqrtCode           string = "from starkware.python.math_utils import isqrt\nn = (ids.n.high << 128) + ids.n.low\nroot = isqrt(n)\nassert 0 <= root < 2 ** 128\nids.root.low = root\nids.root.high = 0"
	uint256MulDivModCode      string = "a = (ids.a.high << 128) + ids.a.low\nb = (ids.b.high << 128) + ids.b.low\ndiv = (ids.div.high << 128) + ids.div.low\nquotient, remainder = divmod(a * b, div)\n\nids.quotient_low.low = quotient & ((1 << 128) - 1)\nids.quotient_low.high = (quotient >> 128) & ((1 << 128) - 1)\nids.quotient_high.low = (quotient >> 256) & ((1 << 128) - 1)\nids.quotient_high.high = quotient >> 384\nids.remainder.low = remainder & ((1 << 128) - 1)\nids.remainder.high = remainder >> 128"

	// ------ Usort hints related code ------
	usortBodyCode                     string = "from collections import defaultdict\n\ninput_ptr = ids.input\ninput_len = int(ids.input_len)\nif __usort_max_size is not None:\n    assert input_len <= __usort_max_size, (\n        f\"usort() can only be used with input_len<={__usort_max_size}. \"\n        f\"Got: input_len={input_len}.\"\n    )\n\npositions_dict = defaultdict(list)\nfor i in range(input_len):\n    val = memory[input_ptr + i]\n    positions_dict[val].append(i)\n\noutput = sorted(positions_dict.keys())\nids.output_len = len(output)\nids.output = segments.gen_arg(output)\nids.multiplicities = segments.gen_arg([len(positions_dict[k]) for k in output])"
	usortEnterScopeCode               string = "vm_enter_scope(dict(__usort_max_size = globals().get('__usort_max_size')))"
	usortVerifyMultiplicityAssertCode string = "assert len(positions) == 0"
	usortVerifyCode                   string = "last_pos = 0\npositions = positions_dict[ids.value][::-1]"
	usortVerifyMultiplicityBodyCode   string = "current_pos = positions.pop()\nids.next_item_index = current_pos - last_pos\nlast_pos = current_pos + 1"

	// ------ Elliptic Curve hints related code ------
	ecNegateCode             string = "from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack\n\ny = pack(ids.point.y, PRIME) % SECP_P\n# The modulo operation in python always returns a nonnegative number.\nvalue = (-y) % SECP_P"
	nondetBigint3V1Code      string = "from starkware.cairo.common.cairo_secp.secp_utils import split\n\nsegments.write_arg(ids.res.address_, split(value))"
	fastEcAddAssignNewXCode  string = "from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack\n\nslope = pack(ids.slope, PRIME)\nx0 = pack(ids.point0.x, PRIME)\nx1 = pack(ids.point1.x, PRIME)\ny0 = pack(ids.point0.y, PRIME)\n\nvalue = new_x = (pow(slope, 2, SECP_P) - x0 - x1) % SECP_P"
	fastEcAddAssignNewYCode  string = "value = new_y = (slope * (x0 - new_x) - y0) % SECP_P"
	ecDoubleSlopeV1Code      string = "from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack\nfrom starkware.python.math_utils import ec_double_slope\n\n# Compute the slope.\nx = pack(ids.point.x, PRIME)\ny = pack(ids.point.y, PRIME)\nvalue = slope = ec_double_slope(point=(x, y), alpha=0, p=SECP_P)"
	reduceV1Code             string = "from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack\n\nvalue = pack(ids.x, PRIME) % SECP_P"
	computeSlopeV1Code       string = "from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack\nfrom starkware.python.math_utils import line_slope\n\n# Compute the slope.\nx0 = pack(ids.point0.x, PRIME)\ny0 = pack(ids.point0.y, PRIME)\nx1 = pack(ids.point1.x, PRIME)\ny1 = pack(ids.point1.y, PRIME)\nvalue = slope = line_slope(point1=(x0, y0), point2=(x1, y1), p=SECP_P)"
	ecDoubleAssignNewXV1Code string = "from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack\n\nslope = pack(ids.slope, PRIME)\nx = pack(ids.point.x, PRIME)\ny = pack(ids.point.y, PRIME)\n\nvalue = new_x = (pow(slope, 2, SECP_P) - 2 * x) % SECP_P"
	ecDoubleAssignNewYV1Code string = "value = new_y = (slope * (x - new_x) - y) % SECP_P"
	ecMulInnerCode           string = "memory[ap] = (ids.scalar % PRIME) % 2"
	isZeroNondetCode         string = "memory[ap] = to_felt_or_relocatable(x == 0)"
	isZeroPackCode           string = "from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack\n\nx = pack(ids.x, PRIME) % SECP_P"
	isZeroDivModCode         string = "from starkware.cairo.common.cairo_secp.secp_utils import SECP_P\nfrom starkware.python.math_utils import div_mod\n\nvalue = x_inv = div_mod(1, x, SECP_P)"
	recoverYCode             string = "from starkware.crypto.signature.signature import ALPHA, BETA, FIELD_PRIME\nfrom starkware.python.math_utils import recover_y\nids.p.x = ids.x\n# This raises an exception if `x` is not on the curve.\nids.p.y = recover_y(ids.x, ALPHA, BETA, FIELD_PRIME)"
	randomEcPointCode        string = "from starkware.crypto.signature.signature import ALPHA, BETA, FIELD_PRIME\nfrom starkware.python.math_utils import random_ec_point\nfrom starkware.python.utils import to_bytes\n\n# Define a seed for random_ec_point that's dependent on all the input, so that:\n#   (1) The added point s is deterministic.\n#   (2) It's hard to choose inputs for which the builtin will fail.\nseed = b\"\".join(map(to_bytes, [ids.p.x, ids.p.y, ids.m, ids.q.x, ids.q.y]))\nids.s.x, ids.s.y = random_ec_point(FIELD_PRIME, ALPHA, BETA, seed)"

	// ------ Signature hints related code ------
	verifyECDSASignatureCode  string = "ecdsa_builtin.add_signature(ids.ecdsa_ptr.address_, (ids.signature_r, ids.signature_s))"
	getPointFromXCode         string = "from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack\n\nx_cube_int = pack(ids.x_cube, PRIME) % SECP_P\ny_square_int = (x_cube_int + ids.BETA) % SECP_P\ny = pow(y_square_int, (SECP_P + 1) // 4, SECP_P)\n\n# We need to decide whether to take y or SECP_P - y.\nif ids.v % 2 == y % 2:\n    value = y\nelse:\n    value = (-y) % SECP_P"
	divModNSafeDivCode        string = "value = k = safe_div(res * b - a, N)"
	importSecp256R1PCode      string = "from starkware.cairo.common.cairo_secp.secp256r1_utils import SECP256R1_P as SECP_P"
	verifyZeroCode            string = "from starkware.cairo.common.cairo_secp.secp_utils import SECP_P, pack\n\nq, r = divmod(pack(ids.val, PRIME), SECP_P)\nassert r == 0, f\"verify_zero: Invalid input {ids.val.d0, ids.val.d1, ids.val.d2}.\"\nids.q = q % PRIME"
	divModNPackedDivmodV1Code string = "from starkware.cairo.common.cairo_secp.secp_utils import N, pack\nfrom starkware.python.math_utils import div_mod, safe_div\n\na = pack(ids.a, PRIME)\nb = pack(ids.b, PRIME)\nvalue = res = div_mod(a, b, N)"

	// ------ Blake Hash hints related code ------
	blake2sAddUint256BigendCode string = "B = 32\nMASK = 2 ** 32 - 1\nsegments.write_arg(ids.data, [(ids.high >> (B * (3 - i))) & MASK for i in range(4)])\nsegments.write_arg(ids.data + 4, [(ids.low >> (B * (3 - i))) & MASK for i in range(4)])"
	blake2sAddUint256Code       string = "B = 32\nMASK = 2 ** 32 - 1\nsegments.write_arg(ids.data, [(ids.low >> (B * i)) & MASK for i in range(4)])\nsegments.write_arg(ids.data + 4, [(ids.high >> (B * i)) & MASK for i in range(4)])"
	blake2sFinalizeCode         string = "# Add dummy pairs of input and output.\nfrom starkware.cairo.common.cairo_blake2s.blake2s_utils import IV, blake2s_compress\n\n_n_packed_instances = int(ids.N_PACKED_INSTANCES)\nassert 0 <= _n_packed_instances < 20\n_blake2s_input_chunk_size_felts = int(ids.INPUT_BLOCK_FELTS)\nassert 0 <= _blake2s_input_chunk_size_felts < 100\n\nmessage = [0] * _blake2s_input_chunk_size_felts\nmodified_iv = [IV[0] ^ 0x01010020] + IV[1:]\noutput = blake2s_compress(\n    message=message,\n    h=modified_iv,\n    t0=0,\n    t1=0,\n    f0=0xffffffff,\n    f1=0,\n)\npadding = (modified_iv + message + [0, 0xffffffff] + output) * (_n_packed_instances - 1)\nsegments.write_arg(ids.blake2s_ptr_end, padding)"
	blake2sFinalizeV2Code       string = "# Add dummy pairs of input and output.\nfrom starkware.cairo.common.cairo_blake2s.blake2s_utils import IV, blake2s_compress\n\n_n_packed_instances = int(ids.N_PACKED_INSTANCES)\nassert 0 <= _n_packed_instances < 20\n_blake2s_input_chunk_size_felts = int(ids.BLAKE2S_INPUT_CHUNK_SIZE_FELTS)\nassert 0 <= _blake2s_input_chunk_size_felts < 100\n\nmessage = [0] * _blake2s_input_chunk_size_felts\nmodified_iv = [IV[0] ^ 0x01010020] + IV[1:]\noutput = blake2s_compress(\n    message=message,\n    h=modified_iv,\n    t0=0,\n    t1=0,\n    f0=0xffffffff,\n    f1=0,\n)\npadding = (modified_iv + message + [0, 0xffffffff] + output) * (_n_packed_instances - 1)\nsegments.write_arg(ids.blake2s_ptr_end, padding)"
	blake2sFinalizeV3Code       string = "# Add dummy pairs of input and output.\nfrom starkware.cairo.common.cairo_blake2s.blake2s_utils import IV, blake2s_compress\n\n_n_packed_instances = int(ids.N_PACKED_INSTANCES)\nassert 0 <= _n_packed_instances < 20\n_blake2s_input_chunk_size_felts = int(ids.BLAKE2S_INPUT_CHUNK_SIZE_FELTS)\nassert 0 <= _blake2s_input_chunk_size_felts < 100\n\nmessage = [0] * _blake2s_input_chunk_size_felts\nmodified_iv = [IV[0] ^ 0x01010020] + IV[1:]\noutput = blake2s_compress(\n    message=message,\n    h=modified_iv,\n    t0=0,\n    t1=0,\n    f0=0xffffffff,\n    f1=0,\n)\npadding = (message + modified_iv + [0, 0xffffffff] + output) * (_n_packed_instances - 1)\nsegments.write_arg(ids.blake2s_ptr_end, padding)"
	blake2sComputeCode          string = "from starkware.cairo.common.cairo_blake2s.blake2s_utils import compute_blake2s_func\ncompute_blake2s_func(segments=segments, output_ptr=ids.output)"

	// ------ Keccak hints related code ------
	unsafeKeccakFinalizeCode         string = "from eth_hash.auto import keccak\nkeccak_input = bytearray()\nn_elms = ids.keccak_state.end_ptr - ids.keccak_state.start_ptr\nfor word in memory.get_range(ids.keccak_state.start_ptr, n_elms):\n    keccak_input += word.to_bytes(16, 'big')\nhashed = keccak(keccak_input)\nids.high = int.from_bytes(hashed[:16], 'big')\nids.low = int.from_bytes(hashed[16:32], 'big')"
	unsafeKeccakCode                 string = "from eth_hash.auto import keccak\n\ndata, length = ids.data, ids.length\n\nif '__keccak_max_size' in globals():\n    assert length <= __keccak_max_size, \\\n        f'unsafe_keccak() can only be used with length<={__keccak_max_size}. ' \\\n        f'Got: length={length}.'\n\nkeccak_input = bytearray()\nfor word_i, byte_i in enumerate(range(0, length, 16)):\n    word = memory[data + word_i]\n    n_bytes = min(16, length - byte_i)\n    assert 0 <= word < 2 ** (8 * n_bytes)\n    keccak_input += word.to_bytes(n_bytes, 'big')\n\nhashed = keccak(keccak_input)\nids.high = int.from_bytes(hashed[:16], 'big')\nids.low = int.from_bytes(hashed[16:32], 'big')"
	cairoKeccakFinalizeCode          string = "# Add dummy pairs of input and output.\n_keccak_state_size_felts = int(ids.KECCAK_STATE_SIZE_FELTS)\n_block_size = int(ids.BLOCK_SIZE)\nassert 0 <= _keccak_state_size_felts < 100\nassert 0 <= _block_size < 10\ninp = [0] * _keccak_state_size_felts\npadding = (inp + keccak_func(inp)) * _block_size\nsegments.write_arg(ids.keccak_ptr_end, padding)"
	keccakWriteArgsCode              string = "segments.write_arg(ids.inputs, [ids.low % 2 ** 64, ids.low // 2 ** 64])\nsegments.write_arg(ids.inputs + 2, [ids.high % 2 ** 64, ids.high // 2 ** 64])"
	blockPermutationCode             string = "from starkware.cairo.common.keccak_utils.keccak_utils import keccak_func\n_keccak_state_size_felts = int(ids.KECCAK_STATE_SIZE_FELTS)\nassert 0 <= _keccak_state_size_felts < 100\n\noutput_values = keccak_func(memory.get_range(\n    ids.keccak_ptr - _keccak_state_size_felts, _keccak_state_size_felts))\nsegments.write_arg(ids.keccak_ptr, output_values)"
	compareBytesInWordCode           string = "memory[ap] = to_felt_or_relocatable(ids.n_bytes < ids.BYTES_IN_WORD)"
	compareKeccakFullRateInBytesCode string = "memory[ap] = to_felt_or_relocatable(ids.n_bytes >= ids.KECCAK_FULL_RATE_IN_BYTES)"
	splitInput3Code                  string = "ids.high3, ids.low3 = divmod(memory[ids.inputs + 3], 256)"
	splitInput6Code                  string = "ids.high6, ids.low6 = divmod(memory[ids.inputs + 6], 256 ** 2)"
	splitInput9Code                  string = "ids.high9, ids.low9 = divmod(memory[ids.inputs + 9], 256 ** 3)"
	splitInput12Code                 string = "ids.high12, ids.low12 = divmod(memory[ids.inputs + 12], 256 ** 4)"
	splitInput15Code                 string = "ids.high15, ids.low15 = divmod(memory[ids.inputs + 15], 256 ** 5)"
	splitOutputMidLowHighCode        string = "tmp, ids.output1_low = divmod(ids.output1, 256 ** 7)\nids.output1_high, ids.output1_mid = divmod(tmp, 2 ** 128)"
	splitOutput0Code                 string = "ids.output0_low = ids.output0 & ((1 << 128) - 1)\nids.output0_high = ids.output0 >> 128"
	SplitNBytesCode                  string = "ids.n_words_to_copy, ids.n_bytes_left = divmod(ids.n_bytes, ids.BYTES_IN_WORD)"

	// ------ Dictionaries hints related code ------
	dictNewCode                           string = "if '__dict_manager' not in globals():\n    from starkware.cairo.common.dict import DictManager\n    __dict_manager = DictManager()\n\nmemory[ap] = __dict_manager.new_dict(segments, initial_dict)\ndel initial_dict"
	defaultDictNewCode                    string = "if '__dict_manager' not in globals():\n    from starkware.cairo.common.dict import DictManager\n    __dict_manager = DictManager()\n\nmemory[ap] = __dict_manager.new_default_dict(segments, ids.default_value)"
	dictReadCode                          string = "dict_tracker = __dict_manager.get_tracker(ids.dict_ptr)\ndict_tracker.current_ptr += ids.DictAccess.SIZE\nids.value = dict_tracker.data[ids.key]"
	dictSquashCopyDictCode                string = "# Prepare arguments for dict_new. In particular, the same dictionary values should be copied\n# to the new (squashed) dictionary.\nvm_enter_scope({\n    # Make __dict_manager accessible.\n    '__dict_manager': __dict_manager,\n    # Create a copy of the dict, in case it changes in the future.\n    'initial_dict': dict(__dict_manager.get_dict(ids.dict_accesses_end)),\n})"
	dictWriteCode                         string = "dict_tracker = __dict_manager.get_tracker(ids.dict_ptr)\ndict_tracker.current_ptr += ids.DictAccess.SIZE\nids.dict_ptr.prev_value = dict_tracker.data[ids.key]\ndict_tracker.data[ids.key] = ids.new_value"
	dictUpdateCode                        string = "# Verify dict pointer and prev value.\ndict_tracker = __dict_manager.get_tracker(ids.dict_ptr)\ncurrent_value = dict_tracker.data[ids.key]\nassert current_value == ids.prev_value, \\\n    f'Wrong previous value in dict. Got {ids.prev_value}, expected {current_value}.'\n\n# Update value.\ndict_tracker.data[ids.key] = ids.new_value\ndict_tracker.current_ptr += ids.DictAccess.SIZE"
	squashDictCode                        string = "dict_access_size = ids.DictAccess.SIZE\naddress = ids.dict_accesses.address_\nassert ids.ptr_diff % dict_access_size == 0, \\\n    'Accesses array size must be divisible by DictAccess.SIZE'\nn_accesses = ids.n_accesses\nif '__squash_dict_max_size' in globals():\n    assert n_accesses <= __squash_dict_max_size, \\\n        f'squash_dict() can only be used with n_accesses<={__squash_dict_max_size}. ' \\\n        f'Got: n_accesses={n_accesses}.'\n# A map from key to the list of indices accessing it.\naccess_indices = {}\nfor i in range(n_accesses):\n    key = memory[address + dict_access_size * i]\n    access_indices.setdefault(key, []).append(i)\n# Descending list of keys.\nkeys = sorted(access_indices.keys(), reverse=True)\n# Are the keys used bigger than range_check bound.\nids.big_keys = 1 if keys[0] >= range_check_builtin.bound else 0\nids.first_key = key = keys.pop()"
	squashDictInnerAssertLenKeysCode      string = "assert len(keys) == 0"
	squashDictInnerCheckAccessIndexCode   string = "new_access_index = current_access_indices.pop()\nids.loop_temps.index_delta_minus1 = new_access_index - current_access_index - 1\ncurrent_access_index = new_access_index"
	squashDictInnerContinueLoopCode       string = "ids.loop_temps.should_continue = 1 if current_access_indices else 0"
	squashDictInnerFirstIterationCode     string = "current_access_indices = sorted(access_indices[key])[::-1]\ncurrent_access_index = current_access_indices.pop()\nmemory[ids.range_check_ptr] = current_access_index"
	squashDictInnerSkipLoopCode           string = "ids.should_skip_loop = 0 if current_access_indices else 1"
	squashDictInnerLenAssertCode          string = "assert len(current_access_indices) == 0"
	squashDictInnerNextKeyCode            string = "assert len(keys) > 0, 'No keys left but remaining_accesses > 0.'\nids.next_key = key = keys.pop()"
	squashDictInnerUsedAccessesAssertCode string = "assert ids.n_used_accesses == len(access_indices[key])"
	dictSquashUpdatePtrCode               string = "# Update the DictTracker's current_ptr to point to the end of the squashed dict.\n__dict_manager.get_tracker(ids.squashed_dict_start).current_ptr = \\\n    ids.squashed_dict_end.address_"

	// ------ Other hints related code ------
	allocSegmentCode          string = "memory[ap] = segments.add()"
	memcpyContinueCopyingCode string = "n -= 1\nids.continue_copying = 1 if n > 0 else 0"
	memsetContinueLoopCode    string = "n -= 1\nids.continue_loop = 1 if n > 0 else 0"
	memcpyEnterScopeCode      string = "vm_enter_scope({'n': ids.len})"
	memsetEnterScopeCode      string = "vm_enter_scope({'n': ids.n})"
	vmEnterScopeCode          string = "vm_enter_scope()"
	vmExitScopeCode           string = "vm_exit_scope()"
	findElementCode           string = "array_ptr = ids.array_ptr\nelm_size = ids.elm_size\nassert isinstance(elm_size, int) and elm_size > 0, \\\n    f'Invalid value for elm_size. Got: {elm_size}.'\nkey = ids.key\n\nif '__find_element_index' in globals():\n    ids.index = __find_element_index\n    found_key = memory[array_ptr + elm_size * __find_element_index]\n    assert found_key == key, \\\n        f'Invalid index found in __find_element_index. index: {__find_element_index}, ' \\\n        f'expected key {key}, found key: {found_key}.'\n    # Delete __find_element_index to make sure it's not used for the next calls.\n    del __find_element_index\nelse:\n    n_elms = ids.n_elms\n    assert isinstance(n_elms, int) and n_elms >= 0, \\\n        f'Invalid value for n_elms. Got: {n_elms}.'\n    if '__find_element_max_size' in globals():\n        assert n_elms <= __find_element_max_size, \\\n            f'find_element() can only be used with n_elms<={__find_element_max_size}. ' \\\n            f'Got: n_elms={n_elms}.'\n\n    for i in range(n_elms):\n        if memory[array_ptr + elm_size * i] == key:\n            ids.index = i\n            break\n    else:\n        raise ValueError(f'Key {key} was not found.')"
	nondetElementsOverTwoCode string = "memory[ap] = to_felt_or_relocatable(ids.elements_end - ids.elements >= 2)"
	nondetElementsOverTenCode string = "memory[ap] = to_felt_or_relocatable(ids.elements_end - ids.elements >= 10)"
	setAddCode                string = "assert ids.elm_size > 0\nassert ids.set_ptr <= ids.set_end_ptr\nelm_list = memory.get_range(ids.elm_ptr, ids.elm_size)\nfor i in range(0, ids.set_end_ptr - ids.set_ptr, ids.elm_size):\n    if memory.get_range(ids.set_ptr + i, ids.elm_size) == elm_list:\n        ids.index = i // ids.elm_size\n        ids.is_elm_in_set = 1\n        break\nelse:\n    ids.is_elm_in_set = 0"
	searchSortedLowerCode     string = "array_ptr = ids.array_ptr\nelm_size = ids.elm_size\nassert isinstance(elm_size, int) and elm_size > 0, \\\n    f'Invalid value for elm_size. Got: {elm_size}.'\n\nn_elms = ids.n_elms\nassert isinstance(n_elms, int) and n_elms >= 0, \\\n    f'Invalid value for n_elms. Got: {n_elms}.'\nif '__find_element_max_size' in globals():\n    assert n_elms <= __find_element_max_size, \\\n        f'find_element() can only be used with n_elms<={__find_element_max_size}. ' \\\n        f'Got: n_elms={n_elms}.'\n\nfor i in range(n_elms):\n    if memory[array_ptr + elm_size * i] >= ids.key:\n        ids.index = i\n        break\nelse:\n    ids.index = n_elms"
)
