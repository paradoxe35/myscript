use std::collections::VecDeque;

use super::SAMPLE_RATE;

/// earshot wants exactly 256 samples (16 ms) at 16 kHz.
const VAD_FRAME: usize = 256;
const VAD_THRESHOLD: f32 = 0.5;

/// Silence this long after speech ends the utterance, so the teleprompter
/// gets text at every pause rather than at the end of the take.
const PAUSE_FRAMES: usize = 31; // ~500 ms
/// Frames kept before onset, recovering the attack the detector needed to fire.
const PREFILL_FRAMES: usize = 28; // ~450 ms
/// Consecutive speech frames before onset is believed, rejecting clicks.
const ONSET_FRAMES: usize = 4;
/// Fewer voiced frames than this is a cough or a click; models hallucinate on it.
const MIN_VOICED_FRAMES: usize = 10; // ~160 ms
/// An utterance is cut here even without a pause: the model window is 30 s
/// and the reader needs text before then.
const MAX_UTTERANCE: usize = SAMPLE_RATE as usize * 20;

const RESAMPLER_CHUNK: usize = 1024;

pub trait Detector: Send {
    fn is_speech(&mut self, frame: &[f32]) -> bool;
}

#[derive(Default)]
pub struct Earshot(earshot::Detector);

impl Detector for Earshot {
    fn is_speech(&mut self, frame: &[f32]) -> bool {
        self.0.predict_f32(frame) >= VAD_THRESHOLD
    }
}

/// Resamples to 16 kHz, then cuts what the detector calls speech into
/// utterances at the pauses.
pub struct Pipeline {
    resampler: Option<rubato::FftFixedIn<f32>>,
    pending: Vec<f32>,
    frame: Vec<f32>,
    detector: Box<dyn Detector>,
    speech: Vec<f32>,
    prefill: VecDeque<Vec<f32>>,
    onset: usize,
    pause: usize,
    voiced: usize,
    in_speech: bool,
    utterances: Vec<Vec<f32>>,
}

impl Pipeline {
    pub fn new(input_rate: u32) -> Self {
        Self::with_detector(input_rate, Box::new(Earshot::default()))
    }

    pub fn with_detector(input_rate: u32, detector: Box<dyn Detector>) -> Self {
        Self {
            resampler: (input_rate != SAMPLE_RATE)
                .then(|| {
                    rubato::FftFixedIn::<f32>::new(
                        input_rate as usize,
                        SAMPLE_RATE as usize,
                        RESAMPLER_CHUNK,
                        1,
                        1,
                    )
                    .ok()
                })
                .flatten(),
            pending: Vec::new(),
            frame: Vec::with_capacity(VAD_FRAME),
            detector,
            speech: Vec::new(),
            prefill: VecDeque::with_capacity(PREFILL_FRAMES),
            onset: 0,
            pause: 0,
            voiced: 0,
            in_speech: false,
            utterances: Vec::new(),
        }
    }

    /// True while an utterance is open, from confirmed onset to the pause that closes it.
    pub fn in_speech(&self) -> bool {
        self.in_speech
    }

    pub fn feed(&mut self, samples: &[f32]) {
        if samples.is_empty() {
            return;
        }
        let resampled = self.resample(samples);
        self.frame_and_classify(resampled);
    }

    /// Utterances closed since the last call, oldest first.
    pub fn take_utterances(&mut self) -> Vec<Vec<f32>> {
        std::mem::take(&mut self.utterances)
    }

    /// Flushes the resampler and closes any open utterance.
    pub fn finish(&mut self) -> Vec<Vec<f32>> {
        if !self.pending.is_empty() {
            let mut tail = std::mem::take(&mut self.pending);
            tail.resize(RESAMPLER_CHUNK, 0.0);
            let flushed = self.resample(&tail);
            self.frame_and_classify(flushed);
        }
        if self.in_speech && !self.frame.is_empty() {
            self.speech.extend_from_slice(&self.frame);
        }
        self.frame.clear();
        if self.in_speech {
            self.close_utterance();
        }
        self.take_utterances()
    }

    fn resample(&mut self, samples: &[f32]) -> Vec<f32> {
        let Some(resampler) = self.resampler.as_mut() else {
            return samples.to_vec();
        };

        use rubato::Resampler;
        self.pending.extend_from_slice(samples);

        let mut out = Vec::new();
        while self.pending.len() >= RESAMPLER_CHUNK {
            let chunk: Vec<f32> = self.pending.drain(..RESAMPLER_CHUNK).collect();
            if let Ok(mut done) = resampler.process(&[chunk], None) {
                out.append(&mut done[0]);
            }
        }
        out
    }

    fn frame_and_classify(&mut self, samples: Vec<f32>) {
        for sample in samples {
            self.frame.push(sample);
            if self.frame.len() == VAD_FRAME {
                let frame = std::mem::replace(&mut self.frame, Vec::with_capacity(VAD_FRAME));
                self.classify(frame);
            }
        }
    }

    fn classify(&mut self, frame: Vec<f32>) {
        let speaking = self.detector.is_speech(&frame);
        self.onset = if speaking { self.onset + 1 } else { 0 };

        if !self.in_speech {
            if self.onset >= ONSET_FRAMES {
                self.in_speech = true;
                self.voiced = ONSET_FRAMES;
                self.speech.extend(self.prefill.drain(..).flatten());
            } else {
                if self.prefill.len() == PREFILL_FRAMES {
                    self.prefill.pop_front();
                }
                self.prefill.push_back(frame);
                return;
            }
        }

        self.speech.extend_from_slice(&frame);
        if speaking {
            self.voiced += 1;
        }
        self.pause = if speaking { 0 } else { self.pause + 1 };

        if self.pause >= PAUSE_FRAMES || self.speech.len() >= MAX_UTTERANCE {
            self.close_utterance();
        }
    }

    fn close_utterance(&mut self) {
        let utterance = std::mem::take(&mut self.speech);
        let voiced = std::mem::take(&mut self.voiced);
        self.in_speech = false;
        self.pause = 0;
        self.onset = 0;
        if voiced >= MIN_VOICED_FRAMES {
            self.utterances.push(utterance);
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Classifies by energy, so segmentation can be tested without earshot's model.
    struct Loud;

    impl Detector for Loud {
        fn is_speech(&mut self, frame: &[f32]) -> bool {
            frame.iter().map(|s| s * s).sum::<f32>() / frame.len() as f32 > 0.01
        }
    }

    fn tone(seconds: f32, rate: u32) -> Vec<f32> {
        let n = (seconds * rate as f32) as usize;
        (0..n)
            .map(|i| (i as f32 * 440.0 * std::f32::consts::TAU / rate as f32).sin() * 0.5)
            .collect()
    }

    fn silence(seconds: f32, rate: u32) -> Vec<f32> {
        vec![0.0; (seconds * rate as f32) as usize]
    }

    fn pipeline() -> Pipeline {
        Pipeline::with_detector(SAMPLE_RATE, Box::new(Loud))
    }

    #[test]
    fn a_pause_closes_the_utterance() {
        let mut p = pipeline();
        p.feed(&tone(1.0, SAMPLE_RATE));
        assert!(p.in_speech());
        assert!(p.take_utterances().is_empty(), "speech is still open");

        p.feed(&silence(0.7, SAMPLE_RATE));
        assert!(!p.in_speech());
        let utterances = p.take_utterances();
        assert_eq!(utterances.len(), 1);

        let seconds = utterances[0].len() as f32 / SAMPLE_RATE as f32;
        assert!(
            (1.3..=1.6).contains(&seconds),
            "1 s of tone plus the pause: {seconds}"
        );
    }

    #[test]
    fn utterances_come_out_in_order_and_short_blips_are_dropped() {
        let mut p = pipeline();
        p.feed(&tone(1.0, SAMPLE_RATE));
        p.feed(&silence(0.7, SAMPLE_RATE));
        p.feed(&tone(0.1, SAMPLE_RATE));
        p.feed(&silence(0.7, SAMPLE_RATE));
        p.feed(&tone(2.0, SAMPLE_RATE));
        p.feed(&silence(0.7, SAMPLE_RATE));

        let utterances = p.take_utterances();
        assert_eq!(utterances.len(), 2, "the 100 ms blip is not an utterance");
        assert!(utterances[0].len() < utterances[1].len());
    }

    #[test]
    fn finish_closes_what_is_still_open() {
        let mut p = pipeline();
        p.feed(&tone(1.0, SAMPLE_RATE));
        let utterances = p.finish();
        assert_eq!(utterances.len(), 1);
        assert!(!p.in_speech());
    }

    #[test]
    fn a_long_utterance_is_cut_without_a_pause() {
        let mut p = pipeline();
        p.feed(&tone(25.0, SAMPLE_RATE));
        let utterances = p.take_utterances();
        assert_eq!(utterances.len(), 1);
        assert_eq!(utterances[0].len(), MAX_UTTERANCE);
        assert!(p.in_speech(), "the rest is still being collected");
    }

    #[test]
    fn silence_alone_yields_nothing() {
        let mut p = pipeline();
        p.feed(&silence(5.0, SAMPLE_RATE));
        assert!(p.finish().is_empty());
    }

    #[test]
    fn input_is_resampled_to_16k() {
        let mut p = Pipeline::with_detector(48_000, Box::new(Loud));
        p.feed(&tone(1.0, 48_000));
        p.feed(&silence(1.0, 48_000));
        let utterances = p.finish();
        assert_eq!(utterances.len(), 1);

        let seconds = utterances[0].len() as f32 / SAMPLE_RATE as f32;
        assert!((1.3..=1.7).contains(&seconds), "{seconds}");
    }
}
