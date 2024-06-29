// Test the AllocSegmentCode Cairo0 hint recognition.

from starkware.cairo.common.alloc import alloc

struct Pair {
  a: felt,
  b: felt,
}

func main() {
  alloc_locals;

  let (local p: Pair*) = alloc();
  assert p.a = 54;
  assert p.b = 77;

  ret;
}
