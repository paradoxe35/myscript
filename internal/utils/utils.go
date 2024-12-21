package utils

import (
	"runtime"

	"github.com/shirou/gopsutil/v4/mem"
)

func GetAvailableRAM() (float64, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}
	return float64(v.Available) / (1024 * 1024 * 1024), nil // Convert to GB
}

func GetCPUCores() int {
	return runtime.NumCPU()
}
