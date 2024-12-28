.PHONY: dev build clean all

all: dev

dev:
	export CGO_CFLAGS_ALLOW="-mfma|-mf16c" && wails dev

build: clean
	export CGO_CFLAGS_ALLOW="-mfma|-mf16c" && wails build -upx -obfuscated -nsis

clean:
	rm -rf build/bin
	rm -rf build/frontend/dist