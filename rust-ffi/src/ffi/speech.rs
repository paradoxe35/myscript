use std::ffi::c_char;
use std::os::raw::c_int;
use std::path::PathBuf;

use crate::ffi::types::{
    Callbacks, FFIErrorCode, SttHandle, c_str_to_string, init_logging, result_to_error_code,
    set_last_error, string_to_c_str,
};
use crate::stt::audio::{self, Recorder};
use crate::stt::host::Host;

pub const SAMPLE_RATE: u32 = crate::stt::SAMPLE_RATE;

pub struct Speech {
    recorder: Recorder,
}

fn speech<'a>(handle: SttHandle) -> Option<&'a Speech> {
    if handle.is_null() {
        set_last_error("Null speech handle provided".to_string());
        return None;
    }
    Some(unsafe { &*(handle as *mut Speech) })
}

/// Every callback is invoked from a library thread.
#[unsafe(no_mangle)]
pub extern "C" fn myscript_stt_new(callbacks: Callbacks) -> SttHandle {
    init_logging();
    let speech = Speech {
        recorder: Recorder::spawn(Host::new(callbacks)),
    };
    Box::into_raw(Box::new(speech)) as SttHandle
}

#[unsafe(no_mangle)]
pub unsafe extern "C" fn myscript_stt_free(handle: SttHandle) {
    if handle.is_null() {
        return;
    }
    let speech = unsafe { Box::from_raw(handle as *mut Speech) };
    speech.recorder.shutdown();
}

/// Loads the model and returns once it is resident, or with the reason it is not.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn myscript_stt_load(handle: SttHandle, path: *const c_char) -> c_int {
    let Some(speech) = speech(handle) else {
        return FFIErrorCode::NullPointer as c_int;
    };
    let path = match unsafe { c_str_to_string(path) } {
        Ok(path) if !path.is_empty() => path,
        Ok(_) => {
            set_last_error("Empty model path provided".to_string());
            return FFIErrorCode::InvalidArgument as c_int;
        }
        Err(e) => {
            set_last_error(format!("Invalid model path: {e}"));
            return FFIErrorCode::InvalidUtf8 as c_int;
        }
    };
    result_to_error_code(speech.recorder.load(PathBuf::from(path)))
}

/// Frees the model once any utterance still transcribing is out.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn myscript_stt_unload(handle: SttHandle) -> c_int {
    let Some(speech) = speech(handle) else {
        return FFIErrorCode::NullPointer as c_int;
    };
    speech.recorder.unload();
    FFIErrorCode::Success as c_int
}

/// Opens the microphone and starts listening. Utterances go to the text
/// callback through the loaded model, or to the audio callback in capture-only mode.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn myscript_stt_start(handle: SttHandle) -> c_int {
    let Some(speech) = speech(handle) else {
        return FFIErrorCode::NullPointer as c_int;
    };
    if speech.recorder.recording() {
        set_last_error("Already recording".to_string());
        return FFIErrorCode::OperationFailed as c_int;
    }
    result_to_error_code(speech.recorder.start())
}

/// Stops listening and returns once the last utterance has been delivered.
/// Blocks for as long as that transcription takes, so call it off the UI thread.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn myscript_stt_stop(handle: SttHandle) -> c_int {
    let Some(speech) = speech(handle) else {
        return FFIErrorCode::NullPointer as c_int;
    };
    result_to_error_code(speech.recorder.stop())
}

/// Stops listening and drops whatever has not been delivered yet.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn myscript_stt_cancel(handle: SttHandle) -> c_int {
    let Some(speech) = speech(handle) else {
        return FFIErrorCode::NullPointer as c_int;
    };
    result_to_error_code(speech.recorder.cancel())
}

#[unsafe(no_mangle)]
pub unsafe extern "C" fn myscript_stt_is_recording(handle: SttHandle) -> bool {
    speech(handle).is_some_and(|s| s.recorder.recording())
}

/// Selects the capture device by name. Null or empty means the system default.
/// Takes effect on the next take.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn myscript_stt_set_device(handle: SttHandle, name: *const c_char) -> c_int {
    let Some(speech) = speech(handle) else {
        return FFIErrorCode::NullPointer as c_int;
    };
    match unsafe { optional(name) } {
        Ok(name) => {
            speech.recorder.set_device(name);
            FFIErrorCode::Success as c_int
        }
        Err(e) => {
            set_last_error(format!("Invalid device name: {e}"));
            FFIErrorCode::InvalidUtf8 as c_int
        }
    }
}

/// Sets the spoken language as an ISO code; null or empty asks the model to
/// detect, which only some can. Takes effect on the next take.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn myscript_stt_set_language(
    handle: SttHandle,
    code: *const c_char,
) -> c_int {
    let Some(speech) = speech(handle) else {
        return FFIErrorCode::NullPointer as c_int;
    };
    match unsafe { optional(code) } {
        Ok(code) => {
            speech.recorder.set_language(code);
            FFIErrorCode::Success as c_int
        }
        Err(e) => {
            set_last_error(format!("Invalid language code: {e}"));
            FFIErrorCode::InvalidUtf8 as c_int
        }
    }
}

/// While on, takes never touch the engine: each utterance reaches the audio
/// callback as PCM for the host to transcribe elsewhere. Takes effect on the next take.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn myscript_stt_set_capture_only(handle: SttHandle, enabled: bool) -> c_int {
    let Some(speech) = speech(handle) else {
        return FFIErrorCode::NullPointer as c_int;
    };
    speech.recorder.set_capture_only(enabled);
    FFIErrorCode::Success as c_int
}

/// Input device names, newline separated, the default marked with a leading '*'.
#[unsafe(no_mangle)]
pub extern "C" fn myscript_stt_devices() -> *mut c_char {
    let (devices, default) = audio::devices();
    string_to_c_str(format_devices(devices, default))
}

pub fn format_devices(devices: Vec<String>, default: Option<String>) -> String {
    devices
        .into_iter()
        .map(|name| {
            if Some(&name) == default.as_ref() {
                format!("*{name}")
            } else {
                name
            }
        })
        .collect::<Vec<_>>()
        .join("\n")
}

/// # Safety
/// `value` must be null or a valid C string.
unsafe fn optional(value: *const c_char) -> Result<Option<String>, &'static str> {
    if value.is_null() {
        return Ok(None);
    }
    let value = unsafe { c_str_to_string(value)? };
    Ok((!value.is_empty()).then_some(value))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn the_default_device_is_starred() {
        let listed = format_devices(
            vec!["Built-in".to_owned(), "USB".to_owned()],
            Some("USB".to_owned()),
        );
        assert_eq!(listed, "Built-in\n*USB");
        assert_eq!(format_devices(vec![], None), "");
    }
}
