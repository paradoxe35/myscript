// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package microphone

import (
	"bytes"
	"encoding/binary"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/gen2brain/malgo"
)

const (
	// If silence is exceeded MAX_SILENCE_TIME, we go stop recording
	MAX_SILENCE_TIME = 1000 * 30 // 30 seconds

	DEFAULT_SAMPLE_RATE = 16000

	DEFAULT_CHANNELS = 1
)

type NoiseConfig struct {
	// Noise detection
	MinDecibels     float64 // Minimum decibel value (-100 default)
	MaxDecibels     float64 // Maximum decibel value (0 default)
	TriggerDecibels float64 // Decibel value to trigger the callback (-30 default)
	NoiseThreshold  float64 // Noise detection sensitivity
	MaxBlankTime    int64   // Maximum time to consider a blank (ms) - 600 default

	// Audio stream
	SampleRate uint32 // Sample rate (16000 default)
	Channels   uint32 // Number of channels (1 default)

	OnSequential func([]byte)           // Callback when silence is detected
	OnStop       func(autoStopped bool) // Callback when recording is stopped
}

type AudioSequencer struct {
	ctx           *malgo.AllocatedContext
	device        *malgo.Device
	config        NoiseConfig
	lastNoiseTime time.Time
	isRecording   bool
	inSpeechModal bool
	mu            sync.Mutex
}

type MicInputDevice struct {
	Name      string
	IsDefault uint32
	ID        malgo.DeviceID
}

func NewAudioSequencer() *AudioSequencer {
	config := NoiseConfig{
		MinDecibels:     -100,
		MaxDecibels:     0,
		TriggerDecibels: -40,
		NoiseThreshold:  -50,
		MaxBlankTime:    500,

		SampleRate: DEFAULT_SAMPLE_RATE,
		Channels:   DEFAULT_CHANNELS,
	}

	return &AudioSequencer{
		config:        config,
		lastNoiseTime: time.Now(),
	}
}

func NewCustomAudioSequencer(config NoiseConfig) *AudioSequencer {
	return &AudioSequencer{
		config:        config,
		lastNoiseTime: time.Now(),
	}
}

func (ar *AudioSequencer) GetDeviceConfig() malgo.DeviceConfig {
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = ar.config.Channels
	deviceConfig.SampleRate = ar.config.SampleRate
	deviceConfig.Alsa.NoMMap = 1

	return deviceConfig
}

func (ar *AudioSequencer) GetNoiseConfig() NoiseConfig {
	return ar.config
}

func (ar *AudioSequencer) GetMicInputDevices() ([]MicInputDevice, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, err
	}

	defer ctx.Free()

	devices, err := ctx.Devices(malgo.Capture)

	if err != nil {
		return nil, err
	}

	var micInputDevices []MicInputDevice

	for _, device := range devices {
		micInputDevices = append(micInputDevices, MicInputDevice{
			ID:        device.ID,
			Name:      device.Name(),
			IsDefault: device.IsDefault,
		})
	}

	return micInputDevices, nil

}

// This function will block until the recording is stopped
func (ar *AudioSequencer) Start(micInputDeviceID []byte) error {
	if ar.isRecording {
		return nil
	}

	micDeviceID, err := ar.convertToDeviceID(micInputDeviceID)
	if err != nil {
		return err
	}

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return err
	}

	ar.ctx = ctx

	var currentBuffer []byte
	ar.lastNoiseTime = time.Now()
	ar.inSpeechModal = false

	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		ar.mu.Lock()
		defer ar.mu.Unlock()

		if !ar.isRecording {
			return
		}

		// Detect noise in the current frame
		hasNoise, db := ar.detectNoise(pSample)

		if hasNoise {
			if !ar.inSpeechModal && db >= ar.config.TriggerDecibels {
				ar.inSpeechModal = true
			}

			currentBuffer = append(currentBuffer, pSample...)
			ar.lastNoiseTime = time.Now()
		} else {
			// Check if silence duration exceeds MaxBlankTime
			if ar.inSpeechModal && time.Since(ar.lastNoiseTime).Milliseconds() > ar.config.MaxBlankTime && len(currentBuffer) > 0 {
				ar.inSpeechModal = false

				// Call the callback with the recorded buffer
				if ar.config.OnSequential != nil {
					// Make a copy of the buffer
					bufferCopy := make([]byte, len(currentBuffer))
					copy(bufferCopy, currentBuffer)
					go ar.config.OnSequential(bufferCopy)
				}
				// Clear the buffer
				currentBuffer = currentBuffer[:0]

				ar.lastNoiseTime = time.Now()
			} else {
				currentBuffer = append(currentBuffer, pSample...)
			}
		}
	}

	deviceConfig := ar.GetDeviceConfig()
	deviceConfig.Capture.DeviceID = micDeviceID.Pointer()

	deviceCallbacks := malgo.DeviceCallbacks{Data: onRecvFrames}

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		return err
	}

	ar.device = device
	ar.isRecording = true

	// Start auto stop
	go ar.autoStop()

	return device.Start()
}

func (ar *AudioSequencer) autoStop() {
	// if silence duration exceeds MaxSilenceTime
	// we stop recording
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !ar.inSpeechModal && time.Since(ar.lastNoiseTime).Milliseconds() > MAX_SILENCE_TIME {
			ar.Stop(true)
			slog.Debug("Auto stop recording after silence", "duration_ms", MAX_SILENCE_TIME)
			break
		} else if !ar.isRecording {
			break
		}
	}
}

func (ar *AudioSequencer) Stop(autoStopped bool) {
	ar.isRecording = false

	if ar.device != nil {
		ar.device.Uninit()
		ar.device = nil
	}

	if ar.ctx != nil {
		ar.ctx.Uninit()
		ar.ctx.Free()
		ar.ctx = nil
	}

	if ar.config.OnStop != nil {
		ar.config.OnStop(autoStopped)
	}
}

func (ar *AudioSequencer) SetStopCallback(callback func(autoStopped bool)) {
	ar.config.OnStop = callback
}

func (ar *AudioSequencer) SetSequentializeCallback(callback func([]byte)) {
	ar.config.OnSequential = callback
}

func (ar *AudioSequencer) Recording() bool {
	return ar.isRecording
}

func (ar *AudioSequencer) detectNoise(samples []byte) (bool, float64) {
	if len(samples) < 2 {
		return false, 0
	}

	var sum float64
	count := len(samples) / 2          // 2 bytes per sample for S16 format
	maxPossibleValue := float64(32768) // Max value for 16-bit audio

	// Calculate RMS (Root Mean Square) amplitude
	for i := 0; i < len(samples); i += 2 {
		sample := math.Abs(float64(int16(binary.LittleEndian.Uint16(samples[i : i+2]))))
		sum += sample * sample
	}

	rms := math.Sqrt(sum / float64(count))
	amplitude := rms / maxPossibleValue

	minDecibels := ar.config.MinDecibels
	maxDecibels := ar.config.MaxDecibels
	noiseThreshold := ar.config.NoiseThreshold

	// Calculate decibels relative to full scale (dBFS)
	db := float64(minDecibels)
	if amplitude > 0 {
		db = 20 * math.Log10(amplitude)
		if db < minDecibels {
			db = minDecibels
		} else if db > maxDecibels {
			db = maxDecibels
		}
	}

	// Return true if the sound is above our noise threshold
	return db > noiseThreshold, db
}

func (ar *AudioSequencer) convertToDeviceID(interfaceData []byte) (malgo.DeviceID, error) {
	// Convert directly to malgo.DeviceID
	var deviceID malgo.DeviceID
	copy(deviceID[:], interfaceData)

	return deviceID, nil
}

func (ar *AudioSequencer) RawBytesToWAV(audioData []byte) ([]byte, error) {
	// Create a buffer to store WAV data
	var wavBuffer bytes.Buffer

	channels := ar.config.Channels
	sampleRate := ar.config.SampleRate

	// Write WAV header
	// RIFF chunk descriptor
	wavBuffer.WriteString("RIFF")                             // ChunkID, 4 bytes
	fileSize := uint32(36 + len(audioData))                   // Total file size - 8 bytes for RIFF header
	binary.Write(&wavBuffer, binary.LittleEndian, fileSize-8) // Size (remaining bytes after this field)
	wavBuffer.WriteString("WAVE")

	// fmt sub-chunk
	wavBuffer.WriteString("fmt ")
	binary.Write(&wavBuffer, binary.LittleEndian, uint32(16)) // Subchunk1Size (16 for PCM)
	binary.Write(&wavBuffer, binary.LittleEndian, uint16(1))  // AudioFormat (1 for PCM)
	binary.Write(&wavBuffer, binary.LittleEndian, uint16(channels))
	binary.Write(&wavBuffer, binary.LittleEndian, sampleRate)
	binary.Write(&wavBuffer, binary.LittleEndian, uint32(sampleRate*uint32(channels)*2)) // ByteRate
	binary.Write(&wavBuffer, binary.LittleEndian, uint16(channels*2))                    // BlockAlign
	binary.Write(&wavBuffer, binary.LittleEndian, uint16(16))                            // BitsPerSample

	// data sub-chunk
	wavBuffer.WriteString("data")
	binary.Write(&wavBuffer, binary.LittleEndian, uint32(len(audioData)))
	wavBuffer.Write(audioData)

	return wavBuffer.Bytes(), nil
}
