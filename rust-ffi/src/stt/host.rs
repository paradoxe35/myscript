use std::ffi::CString;

use crate::ffi::Callbacks;

/// The host's callbacks, wrapped so a panic on the host side ends the call
/// rather than unwinding into the audio or engine thread.
#[derive(Clone, Copy)]
pub struct Host {
    callbacks: Callbacks,
}

impl Host {
    pub fn new(callbacks: Callbacks) -> Self {
        Self { callbacks }
    }

    pub fn text(&self, text: &str) {
        let Ok(text) = CString::new(text) else { return };
        let callback = self.callbacks.text;
        guarded(|| callback(text.as_ptr()));
    }

    pub fn audio(&self, samples: &[f32]) {
        let pcm = pcm16(samples);
        let callback = self.callbacks.audio;
        guarded(|| callback(pcm.as_ptr(), pcm.len()));
    }

    pub fn level(&self, rms: f32) {
        let callback = self.callbacks.level;
        guarded(|| callback(rms));
    }

    pub fn stopped(&self, auto: bool) {
        let callback = self.callbacks.stopped;
        guarded(|| callback(auto));
    }

    pub fn error(&self, message: &str) {
        let Ok(message) = CString::new(message) else { return };
        let callback = self.callbacks.error;
        guarded(|| callback(message.as_ptr()));
    }
}

fn guarded(call: impl FnOnce()) {
    if std::panic::catch_unwind(std::panic::AssertUnwindSafe(call)).is_err() {
        log::error!("host callback panicked");
    }
}

/// Converts samples in [-1.0, 1.0] to 16-bit signed PCM.
pub fn pcm16(samples: &[f32]) -> Vec<i16> {
    samples
        .iter()
        .map(|s| (s.clamp(-1.0, 1.0) * i16::MAX as f32) as i16)
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn pcm16_clamps_and_scales() {
        let pcm = pcm16(&[0.0, 1.0, -1.0, 2.0, 0.5]);
        assert_eq!(pcm, vec![0, i16::MAX, -i16::MAX, i16::MAX, i16::MAX / 2]);
    }
}
