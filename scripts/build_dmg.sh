#!/bin/bash
set -e

ARCH=$1
if [ -z "$ARCH" ]; then
    echo "Usage: $0 <amd64|arm64>"
    exit 1
fi

BINARY_NAME="marubot"
BUILD_DIR="build"
APP_NAME="MaruBot"
APP_BUNDLE="$BUILD_DIR/$APP_NAME.app"
CONTENTS_DIR="$APP_BUNDLE/Contents"
MACOS_DIR="$CONTENTS_DIR/MacOS"
RESOURCES_DIR="$CONTENTS_DIR/Resources"

echo "Creating DMG for macOS $ARCH..."

# 1. Create App Bundle Structure
rm -rf "$APP_BUNDLE"
mkdir -p "$MACOS_DIR"
mkdir -p "$RESOURCES_DIR"

# 2. Copy Binary
cp "$BUILD_DIR/$BINARY_NAME-darwin-$ARCH" "$MACOS_DIR/$BINARY_NAME"
chmod +x "$MACOS_DIR/$BINARY_NAME"

# 3. Create PList
cat > "$CONTENTS_DIR/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>$BINARY_NAME</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>com.marubot.agent</string>
    <key>CFBundleName</key>
    <string>$APP_NAME</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
    <key>LSUIElement</key>
    <string>1</string>
</dict>
</plist>
EOF

# 4. Create DMG
DMG_NAME="$BUILD_DIR/marubot-macos-$ARCH.dmg"
rm -f "$DMG_NAME"

hdiutil create -volname "$APP_NAME" -srcfolder "$APP_BUNDLE" -ov -format UDZO "$DMG_NAME"

echo "✓ Created $DMG_NAME"
