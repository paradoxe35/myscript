use std::ffi::{CStr, CString};
use std::os::raw::{c_char, c_float, c_int, c_void};
use std::sync::Once;

use once_cell::sync::OnceCell;
use parking_lot::Mutex;

pub type SttHandle = *mut c_void;

/// One transcribed utterance, in the order spoken. The string is only valid for
/// the duration of the call.
pub type TextCallback = extern "C" fn(*const c_char);
/// One captured utterance as 16-bit signed PCM, mono, at `SAMPLE_RATE`, for a
/// host that transcribes elsewhere. The buffer is only valid for the call.
pub type AudioCallback = extern "C" fn(*const i16, usize);
/// Microphone level while listening, so the host can draw a meter.
pub type LevelCallback = extern "C" fn(c_float);
/// The take ended; true when it ended by itself after a long silence.
pub type StoppedCallback = extern "C" fn(bool);
/// A transcription failed; listening continues.
pub type ErrorCallback = extern "C" fn(*const c_char);

/// Every callback runs on a library thread, never on the caller's.
#[repr(C)]
#[derive(Clone, Copy)]
pub struct Callbacks {
    pub text: TextCallback,
    pub audio: AudioCallback,
    pub level: LevelCallback,
    pub stopped: StoppedCallback,
    pub error: ErrorCallback,
}

#[repr(C)]
#[derive(Debug, Copy, Clone, PartialEq, Eq)]
pub enum FFIErrorCode {
    Success = 0,
    NullPointer = -1,
    InvalidArgument = -2,
    OperationFailed = -3,
    InvalidUtf8 = -4,
}

static LAST_ERROR: OnceCell<Mutex<Option<String>>> = OnceCell::new();

pub fn set_last_error(err: String) {
    *LAST_ERROR.get_or_init(|| Mutex::new(None)).lock() = Some(err);
}

pub fn take_last_error() -> Option<String> {
    LAST_ERROR.get_or_init(|| Mutex::new(None)).lock().take()
}

/// # Safety
/// `c_str` must be a valid null-terminated C string.
pub unsafe fn c_str_to_string(c_str: *const c_char) -> Result<String, &'static str> {
    unsafe {
        if c_str.is_null() {
            return Err("Null pointer provided");
        }
        CStr::from_ptr(c_str)
            .to_str()
            .map(str::to_owned)
            .map_err(|_| "Invalid UTF-8 in C string")
    }
}

/// Caller must free the result with `myscript_stt_free_string`.
pub fn string_to_c_str(s: String) -> *mut c_char {
    match CString::new(s) {
        Ok(c_string) => c_string.into_raw(),
        Err(_) => {
            set_last_error("String contains null byte".to_string());
            std::ptr::null_mut()
        }
    }
}

pub fn result_to_error_code<T>(result: anyhow::Result<T>) -> c_int {
    match result {
        Ok(_) => FFIErrorCode::Success as c_int,
        Err(e) => {
            set_last_error(format!("{e:#}"));
            FFIErrorCode::OperationFailed as c_int
        }
    }
}

static INIT_LOGGING: Once = Once::new();

pub fn init_logging() {
    INIT_LOGGING.call_once(|| {
        let _ = env_logger::Builder::new()
            .filter_level(log::LevelFilter::Warn)
            .try_init();
        transcribe_cpp::init_logging();
    });
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn last_error_is_taken_once() {
        set_last_error("test error".to_string());
        assert_eq!(take_last_error(), Some("test error".to_string()));
        assert_eq!(take_last_error(), None);
    }

    #[test]
    fn strings_round_trip() {
        let c_str = string_to_c_str("Hello, FFI!".to_string());
        assert!(!c_str.is_null());
        unsafe {
            assert_eq!(c_str_to_string(c_str).unwrap(), "Hello, FFI!");
            let _ = CString::from_raw(c_str);
        }
    }

    #[test]
    fn null_bytes_are_refused() {
        assert!(string_to_c_str("a\0b".to_string()).is_null());
        assert!(take_last_error().is_some());
    }
}
