package utils

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"runtime"
	"time"

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

func TwoByteDataToIntSlice(audioData []byte) []int {
	intData := make([]int, len(audioData)/2)
	for i := 0; i < len(audioData); i += 2 {
		value := int(binary.LittleEndian.Uint16(audioData[i : i+2]))
		intData[i/2] = value
	}
	return intData
}

func MeasureExec(name string) func() {
	start := time.Now()

	return func() {
		log.Printf("%s took %v\n", name, time.Since(start))
	}
}

func B64toBytes(s string) ([]byte, error) {
	var output []byte

	err := json.Unmarshal([]byte(s), &output)
	if err != nil {
		return nil, err
	}

	return output, nil
}
