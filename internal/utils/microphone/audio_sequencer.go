package microphone

import (
	"bytes"
	"encoding/binary"
	"math"
	"sync"
	"time"

	"github.com/gen2brain/malgo"
)

const (
	// If 10 seconds of silence is detected, we go stop recording
	MAX_SILENCE_TIME = 1000 * 20 // 20 seconds
)

type NoiseConfig struct {
	// Noise detection
	MinDecibels    float64 // Minimum decibel value (-100 default)
	maxDecibels    float64 // Maximum decibel value (0 default)
	NoiseThreshold float64 // Noise detection sensitivity
	MaxBlankTime   int64   // Maximum time to consider a blank (ms) - 800 default

	// Audio stream
	SampleRate uint32 // Sample rate (44100 default)
	Channels   uint32 // Number of channels (1 default)

	OnSequential func([]byte)           // Callback when silence is detected
	OnStop       func(autoStopped bool) // Callback when recording is stopped
}

type AudioSequencer struct {
	ctx           *malgo.AllocatedContext
	device        *malgo.Device
	config        NoiseConfig
	lastNoiseTime time.Time
	processing    chan bool
	isRecording   bool
	mu            sync.Mutex
}

func NewAudioSequencer() *AudioSequencer {
	config := NoiseConfig{
		MinDecibels:    -100,
		maxDecibels:    0,
		NoiseThreshold: -35,
		MaxBlankTime:   800,

		SampleRate: 44100,
		Channels:   1,
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

// This function will block until the recording is stopped
func (ar *AudioSequencer) Start(started chan<- bool) error {
	if ar.isRecording {
		return nil
	}

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return err
	}

	ar.ctx = ctx

	var currentBuffer []byte
	// Set this to true to start recording, to prevent unnecessary OnSequential calls on Start
	var triggered = true

	ar.lastNoiseTime = time.Now()

	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		ar.mu.Lock()
		defer ar.mu.Unlock()

		if !ar.isRecording {
			return
		}

		// Detect noise in the current frame
		hasNoise := ar.detectNoise(pSample)

		if hasNoise {
			if triggered {
				triggered = false
			}

			currentBuffer = append(currentBuffer, pSample...)
			ar.lastNoiseTime = time.Now()
		} else {
			// Check if silence duration exceeds MaxBlankTime
			if !triggered && time.Since(ar.lastNoiseTime).Milliseconds() > ar.config.MaxBlankTime && len(currentBuffer) > 0 {
				triggered = true
				// Call the callback with the recorded buffer
				if ar.config.OnSequential != nil {
					// Make a copy of the buffer
					bufferCopy := make([]byte, len(currentBuffer))
					copy(bufferCopy, currentBuffer)
					go ar.config.OnSequential(bufferCopy)
				}
				// Clear the buffer
				currentBuffer = currentBuffer[:0]
			} else {
				currentBuffer = append(currentBuffer, pSample...)
			}
		}

		// if silence duration exceeds MaxSilenceTime
		// we stop recording
		if triggered && time.Since(ar.lastNoiseTime).Milliseconds() > MAX_SILENCE_TIME {
			ar.Stop(true)
		}
	}

	deviceConfig := ar.GetDeviceConfig()
	deviceCallbacks := malgo.DeviceCallbacks{Data: onRecvFrames}

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		return err
	}

	ar.device = device
	ar.isRecording = true

	err = device.Start()
	started <- true

	// Block execution until the device is stopped
	<-ar.processing

	return err
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

	if ar.processing != nil {
		close(ar.processing)
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

func (ar *AudioSequencer) detectNoise(samples []byte) bool {
	if len(samples) < 2 {
		return false
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
	maxDecibels := ar.config.maxDecibels
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
	return db > noiseThreshold
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
