package vm

import (
	"testing"
)

func BenchmarkAssertEqual(b *testing.B) {
	b.Run("assign left to rigth", func(b *testing.B) {
		vm := defaultVirtualMachineWithCode(`
            [ap] = [ap - 1], ap++;
        `)
		writeToDataSegment(vm, 1, 3)
		vm.Context.Ap = 2
		vm.Context.Fp = 2

		noHr := noHintRunner{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := vm.RunStep(&noHr); err != nil {
				b.Error(err)
				break
			}
			vm.Context.Pc.Offset--
		}
	})
	b.Run("assign right to left", func(b *testing.B) {
		vm := defaultVirtualMachineWithCode(`
            [ap - 1] = [ap], ap++;
        `)
		writeToDataSegment(vm, 1, 3)
		vm.Context.Ap = 2
		vm.Context.Fp = 2

		noHr := noHintRunner{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := vm.RunStep(&noHr); err != nil {
				b.Error(err)
				break
			}
			vm.Context.Pc.Offset--
		}
	})
	b.Run("addition", func(b *testing.B) {
		vm := defaultVirtualMachineWithCode(`
            [ap] = [ap - 1] + [ap - 2], ap++;
        `)
		writeToDataSegment(vm, 0, 1)
		writeToDataSegment(vm, 1, 1)
		vm.Context.Ap = 2
		vm.Context.Fp = 2

		noHr := noHintRunner{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := vm.RunStep(&noHr); err != nil {
				b.Error(err)
				break
			}
			vm.Context.Pc.Offset--
		}
	})
	b.Run("substraction", func(b *testing.B) {
		vm := defaultVirtualMachineWithCode(`
            [ap] = [ap - 1] + [ap - 2];
            [ap + 2] = [ap - 1];
            ap += 2;
        `)
		writeToDataSegment(vm, 0, 7)
		writeToDataSegment(vm, 2, 10)
		vm.Context.Ap = 2
		vm.Context.Fp = 2

		noHr := noHintRunner{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := vm.RunStep(&noHr); err != nil {
				b.Error(err)
				break
			}
			vm.Context.Pc.Offset %= 4
		}
	})
	b.Run("multiplication", func(b *testing.B) {
		vm := defaultVirtualMachineWithCode(`
            [ap] = [ap - 1] * [ap - 2] , ap++;
        `)
		writeToDataSegment(vm, 0, 2)
		writeToDataSegment(vm, 1, 3)
		vm.Context.Ap = 2
		vm.Context.Fp = 2

		noHr := noHintRunner{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := vm.RunStep(&noHr); err != nil {
				b.Error(err)
				break
			}
			vm.Context.Pc.Offset--
		}
	})
	b.Run("divition", func(b *testing.B) {
		vm := defaultVirtualMachineWithCode(`
            [ap] = [ap - 1] * [ap - 2];
            [ap + 2] = [ap - 1];
            ap += 2;
        `)
		writeToDataSegment(vm, 0, 4)
		writeToDataSegment(vm, 2, 20)
		vm.Context.Ap = 2
		vm.Context.Fp = 2

		noHr := noHintRunner{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := vm.RunStep(&noHr); err != nil {
				b.Error(err)
				break
			}
			vm.Context.Pc.Offset %= 4
		}
	})
}
