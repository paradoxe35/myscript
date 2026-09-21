use std::sync::mpsc::{Receiver, RecvTimeoutError};
use std::time::{Duration, Instant};

use super::audio::Command;
use super::pipeline::Pipeline;

const DRAIN_INTERVAL: Duration = Duration::from_millis(20);
/// A take with nobody speaking for this long ends by itself.
pub const AUTO_STOP: Duration = Duration::from_secs(30);

/// Lets a take run without a device in tests.
pub trait Source {
    fn rate(&self) -> u32;
    fn take(&self) -> Vec<f32>;
}

/// Where closed utterances go: the engine, or straight to the host.
pub trait Sink {
    fn utterance(&mut self, samples: Vec<f32>);
    /// Called on a stop, after the last utterance; returns once everything
    /// handed over has been delivered.
    fn finish(&mut self);
    /// Called on a cancel; whatever is still queued is dropped.
    fn discard(&mut self);
}

pub enum Outcome {
    Ended,
    Shutdown,
}

/// One take, from Start to Stop, Cancel or the silence timeout.
pub struct Take<S: Source> {
    source: S,
    pipeline: Pipeline,
    auto_stop: Duration,
}

impl<S: Source> Take<S> {
    pub fn new(source: S) -> Self {
        let pipeline = Pipeline::new(source.rate());
        Self {
            source,
            pipeline,
            auto_stop: AUTO_STOP,
        }
    }

    #[cfg(test)]
    pub fn with_pipeline(source: S, pipeline: Pipeline, auto_stop: Duration) -> Self {
        Self {
            source,
            pipeline,
            auto_stop,
        }
    }

    /// Runs until the take ends. `ended` is told whether it ended by itself.
    pub fn run(
        mut self,
        commands: &Receiver<Command>,
        sink: &mut dyn Sink,
        ended: &dyn Fn(bool),
    ) -> Outcome {
        let mut last_speech = Instant::now();

        loop {
            match commands.recv_timeout(DRAIN_INTERVAL) {
                Err(RecvTimeoutError::Timeout) => {}
                Err(RecvTimeoutError::Disconnected) | Ok(Command::Shutdown) => {
                    sink.discard();
                    return Outcome::Shutdown;
                }
                Ok(Command::Stop(reply)) => {
                    self.drain(sink);
                    sink.finish();
                    ended(false);
                    let _ = reply.send(());
                    return Outcome::Ended;
                }
                Ok(Command::Cancel(reply)) => {
                    sink.discard();
                    ended(false);
                    let _ = reply.send(());
                    return Outcome::Ended;
                }
                Ok(Command::Start(reply)) => {
                    let _ = reply.send(Err(anyhow::anyhow!("already recording")));
                }
            }

            self.pull(sink);
            if self.pipeline.in_speech() {
                last_speech = Instant::now();
            } else if last_speech.elapsed() >= self.auto_stop {
                self.drain(sink);
                sink.finish();
                ended(true);
                return Outcome::Ended;
            }
        }
    }

    fn pull(&mut self, sink: &mut dyn Sink) {
        self.pipeline.feed(&self.source.take());
        for utterance in self.pipeline.take_utterances() {
            sink.utterance(utterance);
        }
    }

    fn drain(&mut self, sink: &mut dyn Sink) {
        self.pipeline.feed(&self.source.take());
        for utterance in self.pipeline.finish() {
            sink.utterance(utterance);
        }
    }
}

#[cfg(test)]
mod tests {
    use std::cell::RefCell;
    use std::collections::VecDeque;
    use std::sync::mpsc::{Sender, channel};
    use std::sync::{Arc, Mutex};
    use std::thread;

    use super::*;
    use crate::stt::SAMPLE_RATE;
    use crate::stt::pipeline::Detector;

    struct Loud;

    impl Detector for Loud {
        fn is_speech(&mut self, frame: &[f32]) -> bool {
            frame.iter().map(|s| s * s).sum::<f32>() / frame.len() as f32 > 0.01
        }
    }

    /// Hands out one scripted chunk per drain, then silence forever.
    struct Scripted(Mutex<VecDeque<Vec<f32>>>);

    impl Source for Scripted {
        fn rate(&self) -> u32 {
            SAMPLE_RATE
        }
        fn take(&self) -> Vec<f32> {
            self.0.lock().unwrap().pop_front().unwrap_or_default()
        }
    }

    #[derive(Default)]
    struct Log {
        utterances: Vec<usize>,
        finished: bool,
        discarded: bool,
    }

    struct Collect(Arc<Mutex<Log>>);

    impl Sink for Collect {
        fn utterance(&mut self, samples: Vec<f32>) {
            self.0.lock().unwrap().utterances.push(samples.len());
        }
        fn finish(&mut self) {
            self.0.lock().unwrap().finished = true;
        }
        fn discard(&mut self) {
            self.0.lock().unwrap().discarded = true;
        }
    }

    fn tone(seconds: f32) -> Vec<f32> {
        (0..(seconds * SAMPLE_RATE as f32) as usize)
            .map(|i| (i as f32 * 0.2).sin() * 0.5)
            .collect()
    }

    fn silence(seconds: f32) -> Vec<f32> {
        vec![0.0; (seconds * SAMPLE_RATE as f32) as usize]
    }

    struct Running {
        commands: Sender<Command>,
        log: Arc<Mutex<Log>>,
        ended: Arc<Mutex<Vec<bool>>>,
        handle: thread::JoinHandle<Outcome>,
    }

    fn spawn(script: Vec<Vec<f32>>, auto_stop: Duration) -> Running {
        let (tx, rx) = channel();
        let log = Arc::new(Mutex::new(Log::default()));
        let ended = Arc::new(Mutex::new(Vec::new()));

        let sink_log = log.clone();
        let ended_log = ended.clone();
        let handle = thread::spawn(move || {
            let source = Scripted(Mutex::new(script.into()));
            let pipeline = Pipeline::with_detector(SAMPLE_RATE, Box::new(Loud));
            let mut sink = Collect(sink_log);
            let ended = RefCell::new(ended_log);
            Take::with_pipeline(source, pipeline, auto_stop).run(&rx, &mut sink, &|auto| {
                ended.borrow().lock().unwrap().push(auto)
            })
        });

        Running {
            commands: tx,
            log,
            ended,
            handle,
        }
    }

    fn settle() {
        thread::sleep(DRAIN_INTERVAL * 8);
    }

    #[test]
    fn utterances_are_delivered_in_order_and_stop_finishes() {
        let running = spawn(
            vec![tone(1.0), silence(0.7), tone(2.0), silence(0.7), tone(0.5)],
            AUTO_STOP,
        );
        settle();

        let (reply, done) = channel();
        running.commands.send(Command::Stop(reply)).unwrap();
        done.recv_timeout(Duration::from_secs(2))
            .expect("stop is answered");
        assert!(matches!(running.handle.join().unwrap(), Outcome::Ended));

        let log = running.log.lock().unwrap();
        assert_eq!(
            log.utterances.len(),
            3,
            "two pauses and the stop close three utterances"
        );
        assert!(log.utterances[0] < log.utterances[1]);
        assert!(log.finished);
        assert!(!log.discarded);
        assert_eq!(*running.ended.lock().unwrap(), vec![false]);
    }

    #[test]
    fn cancel_discards_without_finishing() {
        let running = spawn(vec![tone(1.0)], AUTO_STOP);
        settle();

        let (reply, done) = channel();
        running.commands.send(Command::Cancel(reply)).unwrap();
        done.recv_timeout(Duration::from_secs(2))
            .expect("cancel is answered");
        running.handle.join().unwrap();

        let log = running.log.lock().unwrap();
        assert!(log.discarded);
        assert!(!log.finished);
        assert_eq!(*running.ended.lock().unwrap(), vec![false]);
    }

    #[test]
    fn a_long_silence_ends_the_take_by_itself() {
        let running = spawn(vec![tone(1.0), silence(0.7)], Duration::from_millis(150));

        let outcome = running.handle.join().unwrap();
        assert!(matches!(outcome, Outcome::Ended));
        let log = running.log.lock().unwrap();
        assert_eq!(log.utterances.len(), 1);
        assert!(log.finished);
        assert_eq!(*running.ended.lock().unwrap(), vec![true]);
    }

    #[test]
    fn shutdown_ends_the_recorder() {
        let running = spawn(vec![], AUTO_STOP);
        running.commands.send(Command::Shutdown).unwrap();
        assert!(matches!(running.handle.join().unwrap(), Outcome::Shutdown));
        assert!(running.ended.lock().unwrap().is_empty());
    }
}
