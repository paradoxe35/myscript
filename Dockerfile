ARG BASE_IMAGE=ghcr.io/abjrcode/cross-wails:v2.8.2

FROM ${BASE_IMAGE} AS builder

WORKDIR /usr/src/app

# Update Wails
RUN wails update
RUN npm install -g pnpm

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Install Libgtk, webkit
RUN dpkg --add-architecture amd64 \
    && apt-get -qq update \
    && apt-get -qq install -y libsdl2-dev:amd64 libwebkit2gtk-4.1-dev:amd64

RUN dpkg --add-architecture arm64 \
    && apt-get -qq update \
    && apt-get -qq install -y libsdl2-dev:arm64 libwebkit2gtk-4.1-dev:arm64

# Docker injects the value of BUILDARCH into the build process
ARG BUILDARCH
ARG APP_NAME=MyScript

ENV CGO_CFLAGS_ALLOW="-mfma|-mf16c"

# Needed atm due to https://github.com/wailsapp/wails/issues/1921
RUN set -exo pipefail; \
    GOOS=linux GOARCH=amd64 CC=x86_64-linux-gnu-gcc wails build -platform linux/amd64 -skipbindings -tags webkit2_41 -o ${APP_NAME}-amd64; \
    GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc-posix CXX=x86_64-w64-mingw32-g++-posix wails build -skipbindings -s -platform windows/amd64 -webview2 embed -nsis;

ENTRYPOINT [ "/bin/bash" ]

#############################################################

FROM ${BASE_IMAGE}

COPY --from=builder /usr/src/app/build/bin /out

ENTRYPOINT [ "sh", "-c" ]
CMD [ "cp -r /out/. /artifacts/" ]
