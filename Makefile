# Detect operating system
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    EXPORT_CMD := set
    RM_CMD := rm -rf
    MAKE_CMD := mingw32-make.exe
    BUILD_FLAGS := -nsis -webview2 embed
else
    DETECTED_OS := $(shell uname -s)
    EXPORT_CMD := export
    RM_CMD := rm -rf
    MAKE_CMD := make
    ifeq ($(DETECTED_OS),Linux)
        BUILD_FLAGS := -tags webkit2_41 -webview2 embed
    else
        BUILD_FLAGS := -webview2 embed
    endif
endif

.PHONY: dev build clean all lwhisper

all: build

dev:
ifeq ($(DETECTED_OS),Windows)
	set CGO_CFLAGS_ALLOW="-mfma|-mf16c"; wails dev
else
	export CGO_CFLAGS_ALLOW="-mfma|-mf16c" && wails dev
endif

lwhisper:
	@${MAKE_CMD} -C ./whisper libwhisper.a

build: clean lwhisper
ifeq ($(DETECTED_OS),Windows)
	set CGO_CFLAGS_ALLOW="-mfma|-mf16c"; wails build ${BUILD_FLAGS}
else
	export CGO_CFLAGS_ALLOW="-mfma|-mf16c" && wails build ${BUILD_FLAGS}
endif

clean:
	${RM_CMD} build/bin
	${RM_CMD} build/frontend/dist