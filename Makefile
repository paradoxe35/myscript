.PHONY: dev build clean all

all: dev

dev:
	export CGO_CFLAGS_ALLOW="-mfma|-mf16c" && wails dev

build:
	export CGO_CFLAGS_ALLOW="-mfma|-mf16c" && wails build

clean:
	rm -rf build/bin
	rm -rf build/frontend/dist