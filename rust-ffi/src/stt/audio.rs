use std::collections::HashSet;
use std::path::PathBuf;
use std::sync::Arc;
use std::sync::atomic::{AtomicBool, AtomicU64, Ordering};
use std::sync::mpsc::{Receiver, Sender, TryRecvError, channel};
use std::thread;
use std::time::{Duration, Instant};

use anyhow::{Result, anyhow};
use cpal::traits::{DeviceTrait, HostTrait, StreamTrait};
use cpal::{Device, SampleFormat, StreamConfig};
use parking_lot::Mutex;

use super::engine::Engine;
use super::host::Host;
use super::take::{Outcome, Sink, Source, Take};

pub enum Command {
    Start(Sender<Result<()>>),
    Stop(Sender<()>),
    Cancel(Sender<()>),
    Shutdown,
}

/// Read at the start of each take; a change mid-take applies to the next one.
#[derive(Clone, Default)]
struct Settings {
    /// None means the system default.
    device: Option<String>,
    /// None asks the model to detect.
    language: Option<String>,
    /// Hand utterances to the host instead of the engine.
    capture_only: bool,
}

/// Handled on the engine thread in order, so an unload queued after a take's
/// last utterance runs once that utterance is out.
pub enum EngineCommand {
    Load(PathBuf, Sender<Result<()>>),
    Unload,
    Transcribe {
        samples: Vec<f32>,
        language: Option<String>,
        epoch: u64,
    },
    Flush(Sender<()>),
    Shutdown,
}

/// Owns the capture stream on its own thread. cpal delivers audio on a realtime
/// callback that must not block, so it only forwards buffers; every conversion
/// happens here.
pub struct Recorder {
    commands: Sender<Command>,
    engine_commands: Sender<EngineCommand>,
    settings: Arc<Mutex<Settings>>,
    recording: Arc<AtomicBool>,
}

impl Recorder {
    pub fn spawn(host: Host) -> Self {
        let settings = Arc::new(Mutex::new(Settings::default()));
        let recording = Arc::new(AtomicBool::new(false));
        // Bumped by a cancel so utterances still queued for the engine are dropped.
        let epoch = Arc::new(AtomicU64::new(0));

        let (level_tx, level_rx) = channel::<f32>();
        // Levels arrive far faster than a UI can use them; the host is called
        // on this thread, never on the audio callback.
        thread::spawn(move || {
            for rms in level_rx {
                host.level(rms);
            }
        });

        let (engine_tx, engine_rx) = channel();
        let engine_epoch = epoch.clone();
        thread::spawn(move || run_engine(engine_rx, host, engine_epoch));

        let (tx, rx) = channel();
        let recorder = Worker {
            commands: rx,
            levels: level_tx,
            engine_commands: engine_tx.clone(),
            settings: settings.clone(),
            recording: recording.clone(),
            epoch,
            host,
        };
        thread::spawn(move || recorder.run());

        Self {
            commands: tx,
            engine_commands: engine_tx,
            settings,
            recording,
        }
    }

    pub fn set_device(&self, name: Option<String>) {
        self.settings.lock().device = name;
    }

    pub fn set_language(&self, code: Option<String>) {
        self.settings.lock().language = code;
    }

    pub fn set_capture_only(&self, enabled: bool) {
        self.settings.lock().capture_only = enabled;
    }

    pub fn recording(&self) -> bool {
        self.recording.load(Ordering::SeqCst)
    }

    /// Blocks until the model is resident or the load failed.
    pub fn load(&self, path: PathBuf) -> Result<()> {
        let (tx, rx) = channel();
        self.engine_commands
            .send(EngineCommand::Load(path, tx))
            .map_err(|_| anyhow!("engine thread is gone"))?;
        rx.recv().map_err(|_| anyhow!("engine dropped the reply"))?
    }

    /// Queued behind any utterance still transcribing.
    pub fn unload(&self) {
        let _ = self.engine_commands.send(EngineCommand::Unload);
    }

    /// Returns once the microphone is open, or with why it could not be.
    pub fn start(&self) -> Result<()> {
        let (tx, rx) = channel();
        self.commands
            .send(Command::Start(tx))
            .map_err(|_| anyhow!("recorder thread is gone"))?;
        rx.recv()
            .map_err(|_| anyhow!("recorder dropped the reply"))?
    }

    /// Returns once the take has ended and its last utterance is delivered.
    pub fn stop(&self) -> Result<()> {
        let (tx, rx) = channel();
        self.commands
            .send(Command::Stop(tx))
            .map_err(|_| anyhow!("recorder thread is gone"))?;
        rx.recv().map_err(|_| anyhow!("recorder dropped the reply"))
    }

    pub fn cancel(&self) -> Result<()> {
        let (tx, rx) = channel();
        self.commands
            .send(Command::Cancel(tx))
            .map_err(|_| anyhow!("recorder thread is gone"))?;
        rx.recv().map_err(|_| anyhow!("recorder dropped the reply"))
    }

    pub fn shutdown(&self) {
        let _ = self.commands.send(Command::Shutdown);
        let _ = self.engine_commands.send(EngineCommand::Shutdown);
    }
}

struct Worker {
    commands: Receiver<Command>,
    levels: Sender<f32>,
    engine_commands: Sender<EngineCommand>,
    settings: Arc<Mutex<Settings>>,
    recording: Arc<AtomicBool>,
    epoch: Arc<AtomicU64>,
    host: Host,
}

impl Worker {
    fn run(self) {
        loop {
            match self.commands.recv() {
                Ok(Command::Start(reply)) => {
                    if !self.take(reply) {
                        return;
                    }
                }
                Ok(Command::Stop(reply)) | Ok(Command::Cancel(reply)) => {
                    let _ = reply.send(());
                }
                Ok(Command::Shutdown) | Err(_) => return,
            }
        }
    }

    /// Runs one take and returns when it ends, `false` only on shutdown.
    fn take(&self, reply: Sender<Result<()>>) -> bool {
        let Settings {
            device,
            language,
            capture_only,
        } = self.settings.lock().clone();

        let source = match StreamGuard::open(self.levels.clone(), device.as_deref()) {
            Ok(source) => source,
            Err(e) => {
                let _ = reply.send(Err(e));
                return true;
            }
        };
        self.recording.store(true, Ordering::SeqCst);
        let _ = reply.send(Ok(()));

        let mut sink: Box<dyn Sink> = if capture_only {
            Box::new(HostSink { host: self.host })
        } else {
            Box::new(EngineSink {
                commands: self.engine_commands.clone(),
                language,
                epoch: self.epoch.clone(),
            })
        };

        let ended = |auto: bool| {
            self.recording.store(false, Ordering::SeqCst);
            self.host.stopped(auto);
        };
        let outcome = Take::new(source).run(&self.commands, sink.as_mut(), &ended);
        self.recording.store(false, Ordering::SeqCst);
        matches!(outcome, Outcome::Ended)
    }
}

fn run_engine(commands: Receiver<EngineCommand>, host: Host, epoch: Arc<AtomicU64>) {
    let mut engine = Engine::new();
    loop {
        match commands.recv() {
            Ok(EngineCommand::Load(path, reply)) => {
                let _ = reply.send(engine.load(&path));
            }
            Ok(EngineCommand::Unload) => engine.unload(),
            Ok(EngineCommand::Transcribe {
                samples,
                language,
                epoch: taken,
            }) => {
                if taken != epoch.load(Ordering::SeqCst) {
                    continue;
                }
                match engine.transcribe(&samples, language.as_deref()) {
                    Ok(text) if !text.is_empty() => host.text(&text),
                    Ok(_) => {}
                    Err(e) => host.error(&e.to_string()),
                }
            }
            Ok(EngineCommand::Flush(reply)) => {
                let _ = reply.send(());
            }
            Ok(EngineCommand::Shutdown) | Err(_) => return,
        }
    }
}

/// Utterances go to the engine thread, which calls the host once each is transcribed.
struct EngineSink {
    commands: Sender<EngineCommand>,
    language: Option<String>,
    epoch: Arc<AtomicU64>,
}

impl Sink for EngineSink {
    fn utterance(&mut self, samples: Vec<f32>) {
        let _ = self.commands.send(EngineCommand::Transcribe {
            samples,
            language: self.language.clone(),
            epoch: self.epoch.load(Ordering::SeqCst),
        });
    }

    fn finish(&mut self) {
        let (tx, rx) = channel();
        if self.commands.send(EngineCommand::Flush(tx)).is_ok() {
            let _ = rx.recv();
        }
    }

    fn discard(&mut self) {
        self.epoch.fetch_add(1, Ordering::SeqCst);
    }
}

/// Utterances go straight to the host, which transcribes them elsewhere.
struct HostSink {
    host: Host,
}

impl Sink for HostSink {
    fn utterance(&mut self, samples: Vec<f32>) {
        self.host.audio(&samples);
    }
    fn finish(&mut self) {}
    fn discard(&mut self) {}
}

pub struct StreamGuard {
    _stream: cpal::Stream,
    rate: u32,
    channels: usize,
    incoming: Receiver<Vec<f32>>,
}

impl Source for StreamGuard {
    fn rate(&self) -> u32 {
        self.rate
    }

    fn take(&self) -> Vec<f32> {
        let mut out = Vec::new();
        loop {
            match self.incoming.try_recv() {
                Ok(chunk) => out.extend(mono(&chunk, self.channels)),
                Err(TryRecvError::Empty) | Err(TryRecvError::Disconnected) => return out,
            }
        }
    }
}

impl StreamGuard {
    fn open(levels: Sender<f32>, preferred: Option<&str>) -> Result<Self> {
        let device = open_device(preferred)?;
        let config = preferred_config(&device)?;

        let rate = config.config.sample_rate;
        let channels = config.config.channels as usize;
        let (tx, rx) = channel();

        let stream = build_stream(&device, &config, tx, levels)?;
        // cpal 0.18 doesn't auto-start streams; without this the callback never fires.
        stream.play()?;

        Ok(Self {
            _stream: stream,
            rate,
            channels,
            incoming: rx,
        })
    }
}

fn mono(interleaved: &[f32], channels: usize) -> Vec<f32> {
    if channels <= 1 {
        return interleaved.to_vec();
    }
    interleaved
        .chunks(channels)
        .map(|frame| frame.iter().sum::<f32>() / channels as f32)
        .collect()
}

struct SelectedConfig {
    config: StreamConfig,
    format: SampleFormat,
}

/// Falls back to the default device when the chosen one is gone, so an
/// unplugged microphone doesn't stop the reader from working.
fn open_device(preferred: Option<&str>) -> Result<Device> {
    let host = host();

    if let Some(wanted) = preferred {
        match host.input_devices() {
            Ok(mut devices) => {
                if let Some(device) = devices.find(|d| d.to_string() == wanted) {
                    return Ok(device);
                }
                log::warn!("input device '{wanted}' is unavailable, using the default");
            }
            Err(e) => log::warn!("could not enumerate input devices: {e}"),
        }
    }

    host.default_input_device()
        .ok_or_else(|| anyhow!("no input device available"))
}

/// Input device names and the default's name.
pub fn devices() -> (Vec<String>, Option<String>) {
    let host = host();
    let names = host
        .input_devices()
        .map(|devices| {
            let mut seen = HashSet::new();
            devices
                .map(|d| d.to_string())
                .filter(|name| !name.is_empty() && seen.insert(name.clone()))
                .collect()
        })
        .unwrap_or_default();
    let default = host.default_input_device().map(|d| d.to_string());
    (names, default)
}

fn host() -> cpal::Host {
    // ALSA over cpal's default on Linux: PulseAudio/PipeWire both expose an
    // ALSA interface, and going direct avoids a resampling hop.
    #[cfg(target_os = "linux")]
    {
        cpal::host_from_id(cpal::HostId::Alsa).unwrap_or_else(|_| cpal::default_host())
    }
    #[cfg(not(target_os = "linux"))]
    {
        cpal::default_host()
    }
}

/// Uses the device's own rate instead of forcing 16 kHz: forcing a rate the
/// hardware doesn't want can drop Bluetooth headsets into headset profile or
/// make ALSA refuse the stream outright.
fn preferred_config(device: &Device) -> Result<SelectedConfig> {
    let default = device.default_input_config()?;
    let rate = default.sample_rate();

    let best = device
        .supported_input_configs()?
        .filter(|range| range.min_sample_rate() <= rate && rate <= range.max_sample_rate())
        .max_by_key(|range| match range.sample_format() {
            SampleFormat::F32 => 3,
            SampleFormat::I16 => 2,
            SampleFormat::I32 => 1,
            _ => 0,
        });

    match best {
        Some(range) => Ok(SelectedConfig {
            format: range.sample_format(),
            config: range.with_sample_rate(rate).config(),
        }),
        None => Ok(SelectedConfig {
            format: default.sample_format(),
            config: default.config(),
        }),
    }
}

fn build_stream(
    device: &Device,
    selected: &SelectedConfig,
    samples: Sender<Vec<f32>>,
    levels: Sender<f32>,
) -> Result<cpal::Stream> {
    let mut throttle = ErrorThrottle::default();
    let error = move |e: cpal::Error| {
        // The stream recovers on its own, so this is a warning: a glitch to
        // know about when transcription looks off, not a failure.
        if let Some(line) = throttle.record(&e.to_string(), Instant::now()) {
            log::warn!("audio glitch: {line}");
        }
    };

    let stream = match selected.format {
        SampleFormat::F32 => device.build_input_stream(
            selected.config,
            move |data: &[f32], _: &_| forward(data.to_vec(), &samples, &levels),
            error,
            None,
        )?,
        SampleFormat::I16 => device.build_input_stream(
            selected.config,
            move |data: &[i16], _: &_| {
                let converted = data.iter().map(|s| *s as f32 / i16::MAX as f32).collect();
                forward(converted, &samples, &levels)
            },
            error,
            None,
        )?,
        SampleFormat::I32 => device.build_input_stream(
            selected.config,
            move |data: &[i32], _: &_| {
                let converted = data.iter().map(|s| *s as f32 / i32::MAX as f32).collect();
                forward(converted, &samples, &levels)
            },
            error,
            None,
        )?,
        other => return Err(anyhow!("unsupported sample format {other:?}")),
    };

    Ok(stream)
}

/// Runs on the realtime audio callback: send and return, never block.
/// A driver can report a glitch once per audio period, so one bad take can log
/// hundreds of identical lines. Report the first, then how many followed.
const ERROR_SUMMARY_INTERVAL: Duration = Duration::from_secs(5);

#[derive(Default)]
struct ErrorThrottle {
    last: Option<String>,
    repeats: u64,
    since: Option<Instant>,
}

impl ErrorThrottle {
    fn record(&mut self, message: &str, now: Instant) -> Option<String> {
        if self.last.as_deref() != Some(message) {
            self.last = Some(message.to_owned());
            self.repeats = 0;
            self.since = Some(now);
            return Some(message.to_owned());
        }

        self.repeats += 1;

        let elapsed = self.since.map_or(Duration::ZERO, |start| now - start);
        if elapsed < ERROR_SUMMARY_INTERVAL {
            return None;
        }

        let summary = format!(
            "{message} ({} more in the last {}s)",
            self.repeats,
            elapsed.as_secs()
        );
        self.repeats = 0;
        self.since = Some(now);
        Some(summary)
    }
}

fn forward(data: Vec<f32>, samples: &Sender<Vec<f32>>, levels: &Sender<f32>) {
    if !data.is_empty() {
        let sum: f32 = data.iter().map(|s| s * s).sum();
        let _ = levels.send((sum / data.len() as f32).sqrt());
    }
    let _ = samples.send(data);
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::ffi::Callbacks;
    use std::os::raw::{c_char, c_float};

    extern "C" fn text(_: *const c_char) {}
    extern "C" fn audio(_: *const i16, _: usize) {}
    extern "C" fn level(_: c_float) {}
    extern "C" fn stopped(_: bool) {}
    extern "C" fn error(_: *const c_char) {}

    fn recorder() -> Recorder {
        Recorder::spawn(Host::new(Callbacks {
            text,
            audio,
            level,
            stopped,
            error,
        }))
    }

    #[test]
    fn a_failed_load_is_reported_to_the_caller() {
        let recorder = recorder();
        let missing = std::env::temp_dir().join("myscript-nonexistent-model.gguf");

        let message = recorder.load(missing.clone()).unwrap_err().to_string();
        assert!(
            message.contains(&missing.display().to_string()),
            "{message}"
        );

        recorder.shutdown();
    }

    #[test]
    fn stop_and_cancel_without_a_take_are_harmless() {
        let recorder = recorder();
        assert!(!recorder.recording());
        recorder.stop().unwrap();
        recorder.cancel().unwrap();
        recorder.unload();
        recorder.shutdown();
    }

    #[test]
    fn the_first_error_is_reported_and_repeats_are_folded() {
        let mut throttle = ErrorThrottle::default();
        let start = Instant::now();

        assert_eq!(
            throttle.record("underrun", start).as_deref(),
            Some("underrun")
        );
        assert!(throttle.record("underrun", start).is_none());
        assert!(
            throttle
                .record("underrun", start + Duration::from_secs(1))
                .is_none()
        );

        let summary = throttle
            .record("underrun", start + ERROR_SUMMARY_INTERVAL)
            .expect("a summary is due");
        assert!(summary.contains("underrun"), "{summary}");
        assert!(
            summary.contains('3'),
            "the repeats should be counted: {summary}"
        );
    }

    #[test]
    fn a_different_error_is_reported_at_once() {
        let mut throttle = ErrorThrottle::default();
        let start = Instant::now();

        throttle.record("underrun", start);
        assert_eq!(
            throttle.record("device lost", start).as_deref(),
            Some("device lost")
        );
    }

    #[test]
    fn mono_averages_channels() {
        assert_eq!(mono(&[1.0, 0.0, 0.5, 0.5], 2), vec![0.5, 0.5]);
        assert_eq!(mono(&[1.0, 2.0], 1), vec![1.0, 2.0]);
    }
}
