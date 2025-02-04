# Detect operating system
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    EXPORT_CMD := set
    MAKE_CMD := mingw32-make.exe
    BUILD_FLAGS := -nsis -webview2 embed
else
    DETECTED_OS := $(shell uname -s)
    EXPORT_CMD := export
    MAKE_CMD := make
    ifeq ($(DETECTED_OS),Linux)
        BUILD_FLAGS := -tags webkit2_41 -webview2 embed
    else
        BUILD_FLAGS := -webview2 embed
    endif
endif

.PHONY: dev build all cross-env

all: cross-env build

dev:
ifeq ($(DETECTED_OS),Windows)
	@cross-env CGO_CFLAGS_ALLOW="-mfma|-mf16c" wails dev
else
	@rm -rf ~/.cache/MyScript*
	@cross-env CGO_CFLAGS_ALLOW="-mfma|-mf16c" wails dev -tags webkit2_41
endif

build:
ifeq ($(DETECTED_OS),Windows)
	@cross-env CGO_CFLAGS_ALLOW="-mfma|-mf16c" wails build -clean
else
	@cross-env CGO_CFLAGS_ALLOW="-mfma|-mf16c" wails build -clean ${BUILD_FLAGS}
endif

cross-env:
	@npm install -g cross-env
