package stt

import (
	"errors"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeEngine records every command in order, which is what the service's
// contract with Rust is about, and replays what Rust would call back.
type fakeEngine struct {
	mu        sync.Mutex
	log       []string
	callbacks Callbacks
	loadErr   error
	startErr  error
	loading   chan struct{}
	// readyGate blocks the last step before capture, on both the local and the remote path.
	readyGate chan struct{}
}

func (f *fakeEngine) record(entry string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.log = append(f.log, entry)
}

func (f *fakeEngine) entries() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.log)
}

func (f *fakeEngine) SetCallbacks(c Callbacks) { f.callbacks = c }
func (f *fakeEngine) Load(path string) error {
	f.record("load")
	if f.loading != nil {
		<-f.loading
	}
	return f.loadErr
}
func (f *fakeEngine) Unload()                     { f.record("unload") }
func (f *fakeEngine) SetDevice(name string) error { f.record("device " + name); return nil }
func (f *fakeEngine) SetLanguage(code string) error {
	f.record("language " + code)
	f.wait()
	return nil
}
func (f *fakeEngine) SetCaptureOnly(enabled bool) error {
	if enabled {
		f.record("capture-only")
		f.wait()
	} else {
		f.record("transcribe")
	}
	return nil
}
func (f *fakeEngine) wait() {
	if f.readyGate != nil {
		<-f.readyGate
	}
}
func (f *fakeEngine) Start() error { f.record("start"); return f.startErr }
func (f *fakeEngine) Stop() error {
	f.record("stop")
	f.callbacks.Stopped(false)
	return nil
}
func (f *fakeEngine) Cancel() error {
	f.record("cancel")
	f.callbacks.Stopped(false)
	return nil
}
func (f *fakeEngine) Close() { f.record("close") }

func (f *fakeEngine) speak(text string) { f.callbacks.Text(text) }
func (f *fakeEngine) hear(pcm []byte)   { f.callbacks.Audio(pcm) }
func (f *fakeEngine) silence()          { f.callbacks.Stopped(true) }

// heard collects what the listener was told, in order.
type heard struct {
	mu      sync.Mutex
	events  []string
	stopped chan bool
}

func (h *heard) add(event string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, event)
}

func (h *heard) list() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return slices.Clone(h.events)
}

func (h *heard) listener() Listener {
	return Listener{
		State:   func(state State, model Model) { h.add("state " + string(state)) },
		Text:    func(text string) { h.add("text " + text) },
		Error:   func(message string) { h.add("error " + message) },
		Level:   func(rms float32) {},
		Stopped: func(auto bool) { h.add("stopped"); h.stopped <- auto },
	}
}

func (h *heard) waitStopped(t *testing.T) bool {
	t.Helper()
	select {
	case auto := <-h.stopped:
		return auto
	case <-time.After(2 * time.Second):
		t.Fatal("the take never stopped")
		return false
	}
}

func downloadedModel(t *testing.T) (*Store, Model) {
	t.Helper()
	store := NewStore(t.TempDir())
	model := Models().Models[0]

	file, err := os.Create(store.Path(model))
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(model.SizeBytes); err != nil {
		t.Fatal(err)
	}
	file.Close()
	return store, model
}

func newTestService(t *testing.T) (*Service, *fakeEngine, *heard, Options) {
	t.Helper()
	store, model := downloadedModel(t)
	fake := &fakeEngine{}
	h := &heard{stopped: make(chan bool, 4)}
	service := NewService(store, func() (Engine, error) { return fake, nil }, h.listener())
	return service, fake, h, Options{ModelID: model.ID, Language: "en", Device: "USB Mic"}
}

func only(entries []string, prefixes ...string) []string {
	var out []string
	for _, entry := range entries {
		for _, prefix := range prefixes {
			if strings.HasPrefix(entry, prefix) {
				out = append(out, entry)
			}
		}
	}
	return out
}

func TestStartLoadsBeforeListeningAndStopUnloads(t *testing.T) {
	service, fake, h, opts := newTestService(t)

	if err := service.Start(opts); err != nil {
		t.Fatal(err)
	}
	if !service.Recording() {
		t.Fatal("service should be recording after Start")
	}
	if err := service.Stop(); err != nil {
		t.Fatal(err)
	}
	if auto := h.waitStopped(t); auto {
		t.Error("a user stop is not an auto stop")
	}

	want := []string{"load", "transcribe", "language en", "device USB Mic", "start", "stop", "unload"}
	if got := fake.entries(); !slices.Equal(got, want) {
		t.Errorf("engine commands = %v, want %v", got, want)
	}
	states := only(h.list(), "state")
	wantStates := []string{"state loading", "state ready", "state listening", "state idle"}
	if !slices.Equal(states, wantStates) {
		t.Errorf("states = %v, want %v", states, wantStates)
	}
	if service.Recording() {
		t.Error("service still recording after the take ended")
	}
}

func TestCancelUnloadsToo(t *testing.T) {
	service, fake, h, opts := newTestService(t)

	if err := service.Start(opts); err != nil {
		t.Fatal(err)
	}
	if err := service.Cancel(); err != nil {
		t.Fatal(err)
	}
	h.waitStopped(t)

	got := only(fake.entries(), "cancel", "unload")
	if !slices.Equal(got, []string{"cancel", "unload"}) {
		t.Errorf("commands = %v, want cancel then unload", got)
	}
}

func TestUtterancesArriveInOrderBeforeStopped(t *testing.T) {
	service, fake, h, opts := newTestService(t)

	if err := service.Start(opts); err != nil {
		t.Fatal(err)
	}
	fake.speak("one")
	fake.speak("two")
	fake.speak("three")
	service.Stop()
	h.waitStopped(t)

	got := only(h.list(), "text", "stopped")
	want := []string{"text one", "text two", "text three", "stopped"}
	if !slices.Equal(got, want) {
		t.Errorf("delivered %v, want %v", got, want)
	}
}

func TestRemoteUtterancesKeepTheirOrder(t *testing.T) {
	service, fake, h, opts := newTestService(t)

	var mu sync.Mutex
	var wavs [][]byte
	opts.Remote = func(wav []byte, language string) (string, error) {
		mu.Lock()
		wavs = append(wavs, wav)
		mu.Unlock()
		// The first utterance answers last.
		if wav[44] == 1 {
			time.Sleep(100 * time.Millisecond)
		}
		return string(rune('0' + wav[44])), nil
	}

	if err := service.Start(opts); err != nil {
		t.Fatal(err)
	}
	fake.hear([]byte{1, 0})
	fake.hear([]byte{2, 0})
	fake.hear([]byte{3, 0})
	service.Stop()
	h.waitStopped(t)

	got := only(h.list(), "text", "stopped")
	want := []string{"text 1", "text 2", "text 3", "stopped"}
	if !slices.Equal(got, want) {
		t.Errorf("delivered %v, want %v", got, want)
	}
	if !slices.Contains(fake.entries(), "capture-only") || slices.Contains(fake.entries(), "load") {
		t.Errorf("a remote take must capture only, never load: %v", fake.entries())
	}
	if len(wavs) != 3 || string(wavs[0][:4]) != "RIFF" {
		t.Error("utterances should reach the remote transcriber as WAV")
	}
}

func TestRemoteErrorIsSurfacedAndListeningContinues(t *testing.T) {
	service, fake, h, opts := newTestService(t)
	opts.Remote = func([]byte, string) (string, error) { return "", errors.New("quota exceeded") }

	if err := service.Start(opts); err != nil {
		t.Fatal(err)
	}
	fake.hear([]byte{1, 0})
	service.Stop()
	h.waitStopped(t)

	got := only(h.list(), "error", "stopped")
	if !slices.Equal(got, []string{"error quota exceeded", "stopped"}) {
		t.Errorf("delivered %v", got)
	}
}

func TestLoadFailureIsReturnedAndNothingStarts(t *testing.T) {
	service, fake, h, opts := newTestService(t)
	fake.loadErr = errors.New("failed to load model.gguf")

	err := service.Start(opts)
	if err == nil || !strings.Contains(err.Error(), "failed to load") {
		t.Fatalf("Start = %v, want the load failure", err)
	}
	if slices.Contains(fake.entries(), "start") {
		t.Error("capture must not start without a model")
	}
	if service.Recording() {
		t.Error("service must be idle after a failed start")
	}
	states := only(h.list(), "state")
	if !slices.Equal(states, []string{"state loading", "state idle"}) {
		t.Errorf("states = %v", states)
	}
}

func TestStartFailureUnloadsTheModel(t *testing.T) {
	service, fake, _, opts := newTestService(t)
	fake.startErr = errors.New("no input device available")

	if err := service.Start(opts); err == nil {
		t.Fatal("expected the capture failure")
	}
	got := only(fake.entries(), "load", "start", "unload")
	if !slices.Equal(got, []string{"load", "start", "unload"}) {
		t.Errorf("commands = %v, want the model freed after the failed start", got)
	}
}

func TestStopDuringLoadAbandonsTheTake(t *testing.T) {
	service, fake, _, opts := newTestService(t)
	fake.loading = make(chan struct{})

	done := make(chan error, 1)
	go func() { done <- service.Start(opts) }()
	for len(only(fake.entries(), "load")) == 0 {
		time.Sleep(time.Millisecond)
	}

	if err := service.Stop(); err != nil {
		t.Fatal(err)
	}
	close(fake.loading)

	if err := <-done; !errors.Is(err, ErrCancelled) {
		t.Fatalf("Start = %v, want ErrCancelled", err)
	}
	got := only(fake.entries(), "start", "unload")
	if !slices.Equal(got, []string{"unload"}) {
		t.Errorf("commands = %v, want only an unload", got)
	}
}

// A stop that lands after the model is ready but before the microphone opens
// must still end the take, on the local and the remote path alike.
func TestStopBetweenReadyAndStartAbandonsTheTake(t *testing.T) {
	for _, remote := range []bool{false, true} {
		service, fake, h, opts := newTestService(t)
		fake.readyGate = make(chan struct{})
		if remote {
			opts.Remote = func([]byte, string) (string, error) { return "", nil }
		}

		done := make(chan error, 1)
		go func() { done <- service.Start(opts) }()
		gate := "language"
		if remote {
			gate = "capture-only"
		}
		for len(only(fake.entries(), gate)) == 0 {
			time.Sleep(time.Millisecond)
		}
		if remote {
			if err := service.Stop(); err != nil {
				t.Fatal(err)
			}
		} else {
			if !slices.Contains(only(h.list(), "state"), "state ready") {
				t.Fatal("the model should be ready by now")
			}
			if err := service.Stop(); err != nil {
				t.Fatal(err)
			}
		}
		close(fake.readyGate)

		if err := <-done; !errors.Is(err, ErrCancelled) {
			t.Fatalf("remote=%v: Start = %v, want ErrCancelled", remote, err)
		}
		if slices.Contains(fake.entries(), "start") {
			t.Errorf("remote=%v: capture started after the stop", remote)
		}
		if service.Recording() {
			t.Errorf("remote=%v: still recording", remote)
		}
		if !remote && !slices.Contains(fake.entries(), "unload") {
			t.Error("the loaded model was not freed")
		}
	}
}

func TestSilenceEndsTheTakeByItself(t *testing.T) {
	service, fake, h, opts := newTestService(t)

	if err := service.Start(opts); err != nil {
		t.Fatal(err)
	}
	fake.silence()
	if auto := h.waitStopped(t); !auto {
		t.Error("the listener should learn the take stopped by itself")
	}
	if !slices.Contains(fake.entries(), "unload") {
		t.Error("an auto stop must free the model too")
	}
	if service.Recording() {
		t.Error("service still recording after an auto stop")
	}
}

func TestStartRefusesWhatCannotBeLoaded(t *testing.T) {
	service, fake, _, opts := newTestService(t)

	if err := service.Start(Options{}); !errors.Is(err, ErrNoModel) {
		t.Errorf("no model selected = %v, want ErrNoModel", err)
	}
	if err := service.Start(Options{ModelID: "not/in-the-catalogue"}); err == nil {
		t.Error("an unknown model was accepted")
	}
	service.store = NewStore(t.TempDir())
	if err := service.Start(opts); err == nil {
		t.Error("a model that is not downloaded was accepted")
	}
	if len(fake.entries()) != 0 {
		t.Errorf("the engine was touched: %v", fake.entries())
	}
}

func TestSecondStartIsRefusedWhileRecording(t *testing.T) {
	service, _, h, opts := newTestService(t)

	if err := service.Start(opts); err != nil {
		t.Fatal(err)
	}
	if err := service.Start(opts); !errors.Is(err, ErrRecording) {
		t.Errorf("second Start = %v, want ErrRecording", err)
	}
	service.Stop()
	h.waitStopped(t)
}
