// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package utils

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"
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
		slog.Debug(fmt.Sprintf("%s took %v", name, time.Since(start)))
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

func IsDevMode() bool {
	return strings.Contains(os.Args[0], "-dev")
}

func HasInternet() bool {
	client := http.Client{Timeout: time.Duration(5000 * time.Millisecond)}
	if _, err := client.Get("http://clients3.google.com/generate_204"); err != nil {
		return false
	}
	return true
}

func IsAOlderThanBByOneWeek(a, b time.Time) bool {
	duration := b.Sub(a)
	// Check if duration is >= 7 days
	return duration >= 7*24*time.Hour
}

func UniqueStrings(input []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, value := range input {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}
