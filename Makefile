# Detect operating system
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    EXPORT_CMD := set
    RM_CMD := rm -rf
    MAKE_CMD := mingw32-make.exe
else
    DETECTED_OS := $(shell uname -s)
    EXPORT_CMD := export
    RM_CMD := rm -rf
    MAKE_CMD := make
endif

.PHONY: dev build clean all lwhisper

all: dev

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
	set CGO_CFLAGS_ALLOW="-mfma|-mf16c"; wails build -nsis -webview2 embed
else
	export CGO_CFLAGS_ALLOW="-mfma|-mf16c" && wails build -webview2 embed
endif

clean:
	${RM_CMD} build/bin
	${RM_CMD} build/frontend/dist