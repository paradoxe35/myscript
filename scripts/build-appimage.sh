#!/usr/bin/env bash
# Builds a portable AppImage from an already-built Linux binary.
#
# WebKit is deliberately not bundled. A bundled webkit2gtk resolves its helper
# processes through absolute paths that differ across distributions, and the
# bundled GLib then shadows the host's, so the result crashes on any machine
# that is not the build host. Relying on the system webkit2gtk-4.1 matches what
# the deb, rpm and arch packages already require.
set -euo pipefail

ARCH="${1:-amd64}"
VERSION="${VERSION:-0.0.0}"

case "$ARCH" in
  amd64) RUNTIME_ARCH=x86_64 ;;
  arm64) RUNTIME_ARCH=aarch64 ;;
  *) echo "unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY="$ROOT/build/bin/myscript-$ARCH"
OUTPUT="$ROOT/build/bin/myscript-linux-$ARCH.AppImage"

test -x "$BINARY" || { echo "missing binary: $BINARY" >&2; exit 1; }

APPDIR="$(mktemp -d)/AppDir"
trap 'rm -rf "$(dirname "$APPDIR")"' EXIT

mkdir -p "$APPDIR/usr/bin" "$APPDIR/usr/share/applications" \
         "$APPDIR/usr/share/icons/hicolor/256x256/apps" \
         "$APPDIR/usr/share/metainfo"

install -m 0755 "$BINARY" "$APPDIR/usr/bin/myscript"
install -m 0644 "$ROOT/build/appicon.png" "$APPDIR/usr/share/icons/hicolor/256x256/apps/myscript.png"
install -m 0644 "$ROOT/build/linux/myscript.appdata.xml" "$APPDIR/usr/share/metainfo/myscript.appdata.xml"

# Inside an AppImage the binary is not at a fixed path, so Exec must be relative.
sed 's|^Exec=.*|Exec=myscript|' "$ROOT/build/linux/myscript.desktop" \
  > "$APPDIR/usr/share/applications/myscript.desktop"

# appimagetool looks for the desktop entry and icon at the AppDir root.
cp "$APPDIR/usr/share/applications/myscript.desktop" "$APPDIR/myscript.desktop"
cp "$APPDIR/usr/share/icons/hicolor/256x256/apps/myscript.png" "$APPDIR/myscript.png"
ln -sf myscript.png "$APPDIR/.DirIcon"

cat > "$APPDIR/AppRun" <<'RUN'
#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
export PATH="$HERE/usr/bin:$PATH"
export XDG_DATA_DIRS="$HERE/usr/share:${XDG_DATA_DIRS:-/usr/local/share:/usr/share}"
exec "$HERE/usr/bin/myscript" "$@"
RUN
chmod 0755 "$APPDIR/AppRun"

TOOL="$(mktemp -d)/appimagetool"
curl -fsSL -o "$TOOL" \
  "https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-$RUNTIME_ARCH.AppImage"
chmod +x "$TOOL"

# GitHub runners have no FUSE, so the tool has to unpack itself to run.
export APPIMAGE_EXTRACT_AND_RUN=1
ARCH="$RUNTIME_ARCH" "$TOOL" --no-appstream "$APPDIR" "$OUTPUT"

echo "built $OUTPUT"
