// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

// Package permissions asks the operating system for the microphone. Only
// macOS gates it; elsewhere access is always granted.
package permissions

type MicrophoneStatus string

const (
	MicrophoneUndetermined MicrophoneStatus = "undetermined"
	MicrophoneGranted      MicrophoneStatus = "granted"
	MicrophoneDenied       MicrophoneStatus = "denied"
	MicrophoneRestricted   MicrophoneStatus = "restricted"
)
