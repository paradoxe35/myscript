ARG BASE_IMAGE=ghcr.io/paradoxe35/cross-wails:v2.9.2

FROM ${BASE_IMAGE} AS builder

WORKDIR /usr/src/app

# Update Wails and install pnpm
RUN wails update
RUN npm install -g pnpm

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG BUILDARCH
ARG APP_NAME=MyScript

ENV CGO_CFLAGS_ALLOW="-mfma|-mf16c"

# Build for Linux (amd64) and Windows
RUN set -exo pipefail; \
    GOOS=linux GOARCH=amd64 CC=x86_64-linux-gnu-gcc wails build -platform linux/amd64 -skipbindings -tags webkit2_41 -o ${APP_NAME}-amd64; \
    GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc-posix CXX=x86_64-w64-mingw32-g++-posix wails build -skipbindings -s -platform windows/amd64 -webview2 embed -nsis;

ENTRYPOINT [ "/bin/bash" ]

#############################################################

FROM ${BASE_IMAGE}

COPY --from=builder /usr/src/app/build/bin /out

ENTRYPOINT [ "sh", "-c" ]
CMD [ "cp -r /out/. /artifacts/" ]
