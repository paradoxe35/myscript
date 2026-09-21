// Every entry point shares the same raw handle and C string contract.
#![allow(clippy::missing_safety_doc)]

pub mod speech;
pub mod types;

use std::os::raw::c_char;
pub use types::*;

/// Returns the last error as a C string (free with `myscript_stt_free_string`), or NULL if none.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn myscript_stt_get_last_error() -> *const c_char {
    match take_last_error() {
        Some(err) => string_to_c_str(err),
        None => std::ptr::null(),
    }
}

/// Frees a string returned by any function in this crate.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn myscript_stt_free_string(s: *mut c_char) {
    unsafe {
        if !s.is_null() {
            let _ = std::ffi::CString::from_raw(s);
        }
    }
}
