// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package stt

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

var (
	ErrNoModel   = errors.New("no speech model selected; choose one in Settings")
	ErrCancelled = errors.New("recording cancelled while the model was loading")
	ErrRecording = errors.New("already recording")
)

// Reported from the engine's own threads, in the order it happened.
type Callbacks struct {
	Text    func(text string)
	Audio   func(pcm []byte)
	Level   func(rms float32)
	Stopped func(auto bool)
	Error   func(message string)
}

// Load blocks until the model is resident; Stop blocks until the last
// utterance has been delivered.
type Engine interface {
	SetCallbacks(Callbacks)
	Load(path string) error
	Unload()
	SetDevice(name string) error
	SetLanguage(code string) error
	SetCaptureOnly(enabled bool) error
	Start() error
	Stop() error
	Cancel() error
	Close()
}

type State string

const (
	StateIdle      State = "idle"
	StateLoading   State = "loading"
	StateReady     State = "ready"
	StateListening State = "listening"
)

// Every call comes from one goroutine, in order.
type Listener struct {
	State   func(state State, model Model)
	Text    func(text string)
	Error   func(message string)
	Level   func(rms float32)
	Stopped func(auto bool)
}

// Turns a WAV utterance into text elsewhere, such as a remote API.
type Transcriber func(wav []byte, language string) (string, error)

type Options struct {
	// ModelID selects the local model; ignored when Remote is set.
	ModelID string
	// Language is an ISO code, empty to let the model detect.
	Language string
	// Device is the microphone name, empty for the system default.
	Device string
	// Remote, when set, receives each utterance instead of the local engine.
	Remote Transcriber
}

// Runs one take at a time. The model is loaded for the take and freed when it
// ends, so memory is only held while the user is reading.
type Service struct {
	store     *Store
	newEngine func() (Engine, error)
	listener  Listener

	mu        sync.Mutex
	engine    Engine
	phase     State
	loaded    bool
	cancelled bool
	remote    Transcriber
	language  string
	lastLevel time.Time

	results chan *result
}

// One slot in the delivery order: a transcript, a remote request still in
// flight, or the end of the take.
type result struct {
	text    string
	err     error
	done    chan struct{}
	stopped *bool
}

const levelEvery = 66 * time.Millisecond

func NewService(store *Store, newEngine func() (Engine, error), listener Listener) *Service {
	s := &Service{
		store:     store,
		newEngine: newEngine,
		listener:  listener,
		phase:     StateIdle,
		results:   make(chan *result, 64),
	}
	go s.deliver()
	return s
}

func (s *Service) Store() *Store { return s.store }

func (s *Service) Recording() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.phase != StateIdle
}

func (s *Service) ensureEngine() (Engine, error) {
	if s.engine != nil {
		return s.engine, nil
	}
	engine, err := s.newEngine()
	if err != nil {
		return nil, err
	}
	engine.SetCallbacks(Callbacks{
		Text:    s.onText,
		Audio:   s.onAudio,
		Level:   s.onLevel,
		Stopped: s.onStopped,
		Error:   s.onError,
	})
	s.engine = engine
	return engine, nil
}

// Blocks through the model load and returns once the microphone is open.
func (s *Service) Start(opts Options) error {
	s.mu.Lock()
	if s.phase != StateIdle {
		s.mu.Unlock()
		return ErrRecording
	}
	engine, err := s.ensureEngine()
	if err != nil {
		s.mu.Unlock()
		return err
	}

	var model Model
	var path string
	if opts.Remote == nil {
		if model, path, err = s.modelPath(opts.ModelID); err != nil {
			s.mu.Unlock()
			return err
		}
	}
	s.phase = StateLoading
	s.cancelled = false
	s.remote = opts.Remote
	s.language = opts.Language
	s.mu.Unlock()

	if opts.Remote == nil {
		if err := s.load(engine, model, path, opts.Language); err != nil {
			return err
		}
	} else if err := engine.SetCaptureOnly(true); err != nil {
		s.settle(engine)
		return err
	}

	if err := engine.SetDevice(opts.Device); err != nil {
		slog.Warn("Could not select the microphone", "device", opts.Device, "error", err)
	}

	// Held through Start so a Stop landing now waits and then sees a listening take.
	s.mu.Lock()
	if s.cancelled {
		s.mu.Unlock()
		s.settle(engine)
		return ErrCancelled
	}
	if err := engine.Start(); err != nil {
		s.mu.Unlock()
		s.settle(engine)
		return err
	}
	s.phase = StateListening
	s.mu.Unlock()

	s.listener.State(StateListening, model)
	return nil
}

func (s *Service) load(engine Engine, model Model, path, language string) error {
	s.listener.State(StateLoading, model)
	if err := engine.Load(path); err != nil {
		s.settle(engine)
		return err
	}

	s.mu.Lock()
	cancelled := s.cancelled
	s.loaded = true
	s.mu.Unlock()
	if cancelled {
		s.settle(engine)
		return ErrCancelled
	}
	s.listener.State(StateReady, model)

	if err := engine.SetCaptureOnly(false); err != nil {
		s.settle(engine)
		return err
	}
	if err := engine.SetLanguage(model.TranscribeLanguage(language)); err != nil {
		s.settle(engine)
		return err
	}
	return nil
}

// Returns to idle after a start that did not reach listening.
func (s *Service) settle(engine Engine) {
	s.mu.Lock()
	loaded := s.loaded
	s.loaded = false
	s.phase = StateIdle
	s.mu.Unlock()

	if loaded {
		engine.Unload()
	}
	s.listener.State(StateIdle, Model{})
}

func (s *Service) modelPath(id string) (Model, string, error) {
	if id == "" {
		return Model{}, "", ErrNoModel
	}
	model, ok := FindModel(id)
	if !ok {
		return Model{}, "", fmt.Errorf("unknown model %q", id)
	}
	if !s.store.Downloaded(model) {
		return Model{}, "", fmt.Errorf("%s is not downloaded yet", model.Name)
	}
	return model, s.store.Path(model), nil
}

// The engine's stopped callback finishes the take. A take still loading its
// model is abandoned once the load returns.
func (s *Service) Stop() error {
	return s.end(Engine.Stop)
}

// Ends the take and drops whatever has not been delivered yet.
func (s *Service) Cancel() error {
	return s.end(Engine.Cancel)
}

func (s *Service) end(command func(Engine) error) error {
	s.mu.Lock()
	engine, phase := s.engine, s.phase
	if phase == StateLoading {
		s.cancelled = true
	}
	s.mu.Unlock()

	if engine == nil || phase != StateListening {
		return nil
	}
	return command(engine)
}

func (s *Service) Close() {
	s.mu.Lock()
	engine := s.engine
	s.engine = nil
	s.phase = StateIdle
	s.loaded = false
	s.mu.Unlock()

	if engine != nil {
		engine.Close()
	}
}

func (s *Service) onText(text string) {
	done := make(chan struct{})
	close(done)
	s.results <- &result{text: text, done: done}
}

// Sends the utterance out at once and reserves its place in the order, so a
// slow request never lets a later one overtake it.
func (s *Service) onAudio(pcm []byte) {
	s.mu.Lock()
	remote, language := s.remote, s.language
	s.mu.Unlock()
	if remote == nil {
		return
	}

	r := &result{done: make(chan struct{})}
	go func() {
		r.text, r.err = remote(WAV(pcm), language)
		close(r.done)
	}()
	s.results <- r
}

func (s *Service) onLevel(rms float32) {
	if s.listener.Level == nil {
		return
	}
	s.mu.Lock()
	due := time.Since(s.lastLevel) >= levelEvery
	if due {
		s.lastLevel = time.Now()
	}
	s.mu.Unlock()
	if due {
		s.listener.Level(rms)
	}
}

func (s *Service) onStopped(auto bool) {
	s.results <- &result{stopped: &auto}
}

func (s *Service) onError(message string) {
	done := make(chan struct{})
	close(done)
	s.results <- &result{err: errors.New(message), done: done}
}

func (s *Service) deliver() {
	for r := range s.results {
		if r.stopped != nil {
			s.finish(*r.stopped)
			continue
		}
		<-r.done
		switch {
		case r.err != nil:
			s.listener.Error(r.err.Error())
		case r.text != "":
			s.listener.Text(r.text)
		}
	}
}

// Runs once every utterance of the take is out, so the unload never races a
// transcription and the host hears "stopped" last.
func (s *Service) finish(auto bool) {
	s.mu.Lock()
	engine := s.engine
	loaded := s.loaded
	s.loaded = false
	s.phase = StateIdle
	s.remote = nil
	s.mu.Unlock()

	if loaded && engine != nil {
		engine.Unload()
		slog.Info("Speech model unloaded")
	}
	s.listener.Stopped(auto)
	s.listener.State(StateIdle, Model{})
}
