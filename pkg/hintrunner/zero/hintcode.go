package zero

const (
	// This is a block for hint code strings where there is a single
	// hint per function it belongs to (with some exceptions like testAssignCode).
	allocSegmentCode string = "memory[ap] = segments.add()"
	isLeFeltCode     string = "memory[ap] = 0 if (ids.a % PRIME) <= (ids.b % PRIME) else 1"
	assertLtFeltCode string = "from starkware.cairo.common.math_utils import assert_integer\nassert_integer(ids.a)\nassert_integer(ids.b)\nassert (ids.a % PRIME) < (ids.b % PRIME), \\\n    f'a = {ids.a % PRIME} is not less than b = {ids.b % PRIME}.'"

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

	// Uint256 hints related code
	Uint256AddCode = "sum_low = ids.a.low + ids.b.low\nids.carry_low = 1 if sum_low >= ids.SHIFT else 0\nsum_high = ids.a.high + ids.b.high + ids.carry_low\nids.carry_high = 1 if sum_high >= ids.SHIFT else 0"
)
