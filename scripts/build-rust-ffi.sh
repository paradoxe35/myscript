#!/usr/bin/env bash
# Build the Rust speech library and copy it into lib/. Usage: build-rust-ffi.sh [target-triple]
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root/rust-ffi"

target="${1:-}"
cargo_args=(build --release)
if [ -n "$target" ]; then
  cargo_args+=(--target "$target")
  out_dir="target/$target/release"
else
  out_dir="target/release"
fi

# ggml defaults to -march=native, which would tie a release to the build
# machine. Pin an AVX2 baseline (x86 from ~2013); inert on arm64.
# TRANSCRIBE_USE_SYSTEM_BLAS=OFF: the decoder would otherwise call cblas when a
# BLAS is found at build time, which the final cgo link never provides.
export TRANSCRIBE_CMAKE_ARGS="${TRANSCRIBE_CMAKE_ARGS:--DGGML_NATIVE=OFF -DGGML_SSE42=ON -DGGML_AVX=ON -DGGML_AVX2=ON -DGGML_FMA=ON -DGGML_F16C=ON -DGGML_BMI2=ON -DTRANSCRIBE_USE_SYSTEM_BLAS=OFF}"

# ggml clears CMAKE_STATIC_LIBRARY_PREFIX on WIN32, so it installs ggml.a while
# its link manifest says "ggml" and rustc looks for libggml.a.
normalize_native_archives() {
  shopt -s nullglob
  local archive name dir
  for archive in "$out_dir"/build/transcribe-cpp-sys-*/out/lib/*.a; do
    name="$(basename "$archive")"
    dir="$(dirname "$archive")"
    case "$name" in
      lib*) continue ;;
    esac
    cp -f "$archive" "$dir/lib$name"
    echo "  aliased $name -> lib$name"
  done
}

if ! cargo "${cargo_args[@]}"; then
  # The retry reuses the cached build-script output, so nothing native is rebuilt.
  echo "Rust build failed; normalizing native archive names and retrying..."
  normalize_native_archives
  cargo "${cargo_args[@]}"
fi

mkdir -p "$repo_root/lib"
cp "$out_dir/libmyscript_stt.a" "$repo_root/lib/"
echo "lib/libmyscript_stt.a is ready"
