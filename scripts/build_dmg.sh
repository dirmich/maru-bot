#!/bin/bash
set -e

ARCH=$1
if [ -z "$ARCH" ]; then
    echo "Usage: $0 <amd64|arm64>"
    exit 1
fi

BINARY_NAME="marubot"
BUILD_DIR="build"

# 1. Version Detection
VERSION=$(grep '^const Version =' pkg/config/version.go | cut -d '"' -f 2)
if [ -z "$VERSION" ]; then VERSION="0.0.0"; fi

echo "Creating DMG for macOS $ARCH (v$VERSION)..."

# 2. Create App Bundle Structure
APP_NAME="MaruBot"
APP_BUNDLE="$BUILD_DIR/$APP_NAME.app"
CONTENTS_DIR="$APP_BUNDLE/Contents"
MACOS_DIR="$CONTENTS_DIR/MacOS"
RESOURCES_DIR="$CONTENTS_DIR/Resources"

rm -rf "$APP_BUNDLE"
mkdir -p "$MACOS_DIR"
mkdir -p "$RESOURCES_DIR"

# 3. Copy Binary
cp "$BUILD_DIR/$BINARY_NAME-darwin-$ARCH" "$MACOS_DIR/$BINARY_NAME"
chmod +x "$MACOS_DIR/$BINARY_NAME"

# 4. Copy Icon (Try to use PNG if available, macOS prefers .icns but PNG works in some contexts or we can just leave it)
if [ -f "cmd/marubot/assets/tray_icon.png" ]; then
    cp "cmd/marubot/assets/tray_icon.png" "$RESOURCES_DIR/AppIcon.png"
fi

# 5. Create PList
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
    <string>$VERSION</string>
    <key>LSUIElement</key>
    <string>1</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF

# 6. Prepare DMG Root (THIS IS THE FIX)
# We need a folder that contains the .app AND a symlink to Applications
DMG_ROOT="$BUILD_DIR/dmg-root-$ARCH"
rm -rf "$DMG_ROOT"
mkdir -p "$DMG_ROOT"

# Copy the .app bundle into the DMG root
cp -R "$APP_BUNDLE" "$DMG_ROOT/"

# Create symlink to Applications
ln -s /Applications "$DMG_ROOT/Applications"

# 7. Create DMG from the Root Folder
DMG_NAME="$BUILD_DIR/marubot-macos-$ARCH.dmg"
rm -f "$DMG_NAME"

hdiutil create -volname "$APP_NAME v$VERSION" -srcfolder "$DMG_ROOT" -ov -format UDZO "$DMG_NAME"

# Cleanup
rm -rf "$DMG_ROOT"

echo "✓ Created $DMG_NAME with proper installation structure."
