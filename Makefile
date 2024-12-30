.PHONY: dev build clean all lwhisper

all: dev

dev:
	export CGO_CFLAGS_ALLOW="-mfma|-mf16c" && wails dev

lwhisper:
	@${MAKE} -C ./whisper libwhisper.a

build: clean lwhisper
	export CGO_CFLAGS_ALLOW="-mfma|-mf16c" && wails build -nsis -webview2 embed

clean:
	rm -rf build/bin
	rm -rf build/frontend/dist