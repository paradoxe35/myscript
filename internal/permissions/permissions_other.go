// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

//go:build !darwin

package permissions

// Linux and Windows hand out the microphone without asking the app.
func RequestMicrophone() MicrophoneStatus { return MicrophoneGranted }

func OpenMicrophoneSettings() {}
