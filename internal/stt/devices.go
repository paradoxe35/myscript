// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package stt

import "strings"

// Device is a microphone the host can record from.
type Device struct {
	Name      string
	IsDefault bool
}

// ParseDevices reads the newline-separated list the FFI returns, where the
// default carries a leading '*'.
func ParseDevices(listed string) []Device {
	if strings.TrimSpace(listed) == "" {
		return nil
	}

	lines := strings.Split(listed, "\n")
	devices := make([]Device, 0, len(lines))
	seen := make(map[string]int, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		name, isDefault := strings.CutPrefix(line, "*")
		if index, exists := seen[name]; exists {
			devices[index].IsDefault = devices[index].IsDefault || isDefault
			continue
		}
		seen[name] = len(devices)
		devices = append(devices, Device{Name: name, IsDefault: isDefault})
	}
	return devices
}
