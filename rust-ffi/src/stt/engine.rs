use std::path::{Path, PathBuf};

use anyhow::{Result, anyhow};
use transcribe_cpp::RunOptions;

/// Holds the loaded model between utterances. Loading costs seconds and each
/// utterance costs milliseconds, so the session stays resident for a take.
#[derive(Default)]
pub struct Engine {
    loaded: Option<Loaded>,
}

struct Loaded {
    path: PathBuf,
    session: transcribe_cpp::Session,
}

impl Engine {
    pub fn new() -> Self {
        Self { loaded: None }
    }

    pub fn loaded(&self) -> bool {
        self.loaded.is_some()
    }

    pub fn unload(&mut self) {
        self.loaded = None;
    }

    /// Idempotent for the resident model. The old model is freed before the new
    /// one is read, so two are never in memory at once.
    pub fn load(&mut self, path: &Path) -> Result<()> {
        if self.loaded.as_ref().is_some_and(|l| l.path == path) {
            return Ok(());
        }
        self.loaded = None;

        let model =
            transcribe_cpp::Model::load_with(path, &transcribe_cpp::ModelOptions::default())
                .map_err(|e| anyhow!("failed to load {}: {e}", path.display()))?;
        let session = model
            .session_with(&transcribe_cpp::SessionOptions::default())
            .map_err(|e| anyhow!("failed to open a session: {e}"))?;

        self.loaded = Some(Loaded {
            path: path.to_path_buf(),
            session,
        });
        Ok(())
    }

    pub fn transcribe(&mut self, samples: &[f32], language: Option<&str>) -> Result<String> {
        let loaded = self
            .loaded
            .as_mut()
            .ok_or_else(|| anyhow!("no model loaded"))?;

        let options = RunOptions {
            language: language.map(str::to_owned),
            ..Default::default()
        };

        loaded
            .session
            .run(samples, &options)
            .map(|out| out.text.trim().to_owned())
            .map_err(|e| anyhow!("transcription failed: {e}"))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn a_missing_model_fails_by_name() {
        let missing = std::env::temp_dir().join("myscript-missing-model.gguf");
        let mut engine = Engine::new();

        let message = engine.load(&missing).expect_err("must fail").to_string();
        assert!(
            message.contains(&missing.display().to_string()),
            "{message}"
        );
        assert!(!engine.loaded());
    }

    #[test]
    fn transcribing_without_a_model_fails() {
        let mut engine = Engine::new();
        let err = engine.transcribe(&[0.0; 1600], None).unwrap_err();
        assert!(err.to_string().contains("no model loaded"));
    }
}
