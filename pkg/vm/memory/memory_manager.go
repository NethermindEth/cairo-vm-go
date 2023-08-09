package memory

type MemoryManager struct {
	Memory *Memory
}

func CreateMemoryManager() (*MemoryManager, error) {
	memory := InitializeEmptyMemory()

	return &MemoryManager{
		Memory: memory,
	}, nil
}
