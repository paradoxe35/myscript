# Detect operating system
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    RUST_TARGET := x86_64-pc-windows-gnu
    BUILD_FLAGS := -nsis -webview2 embed
else
    DETECTED_OS := $(shell uname -s)
    ifeq ($(DETECTED_OS),Linux)
        RUST_TARGET :=
        BUILD_FLAGS := -tags webkit2_41 -webview2 embed
    else
        RUST_TARGET :=
        BUILD_FLAGS := -webview2 embed
    endif
endif

RUST_LIB := lib/libmyscript_stt.a
RUST_BUILD := bash scripts/build-rust-ffi.sh
RUST_SOURCES := rust-ffi/Cargo.toml rust-ffi/Cargo.lock rust-ffi/build.rs rust-ffi/cbindgen.toml $(shell find rust-ffi/src -type f)

.PHONY: all dev build build-rust build-rust-linux build-rust-windows build-rust-darwin-amd64 build-rust-darwin-arm64 build-frontend test test-rust test-go test-frontend clean-rust

all: build

# The speech library is linked by cgo; wails cannot run without it.
$(RUST_LIB): $(RUST_SOURCES)
	$(MAKE) build-rust

build-rust:
	$(RUST_BUILD) $(RUST_TARGET)

build-rust-linux:
	$(RUST_BUILD)

build-rust-windows:
	rustup target add x86_64-pc-windows-gnu
	$(RUST_BUILD) x86_64-pc-windows-gnu

build-rust-darwin-amd64:
	rustup target add x86_64-apple-darwin
	$(RUST_BUILD) x86_64-apple-darwin

build-rust-darwin-arm64:
	rustup target add aarch64-apple-darwin
	$(RUST_BUILD) aarch64-apple-darwin

dev: $(RUST_LIB)
ifeq ($(DETECTED_OS),Windows)
	wails dev
else
	rm -rf ~/.cache/MyScript*
	wails dev -tags webkit2_41
endif

build: $(RUST_LIB)
	wails build -clean $(BUILD_FLAGS)

# main.go embeds frontend/dist, so every go command needs the frontend built first.
build-frontend:
	cd frontend && pnpm install --frozen-lockfile && pnpm run build

test: test-rust test-frontend test-go

test-rust:
	cd rust-ffi && cargo test --release

test-go: $(RUST_LIB) build-frontend
	go vet ./...
	go test ./...

test-frontend: build-frontend
	cd frontend && pnpm test

clean-rust:
	cd rust-ffi && cargo clean
	rm -f $(RUST_LIB)
