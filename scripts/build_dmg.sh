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
VERSION=$(grep 'const Version =' pkg/config/version.go | head -n 1 | cut -d '"' -f 2)
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

# 4. Copy Icon
if [ -f "cmd/marubot/assets/app_icon.png" ]; then
    cp "cmd/marubot/assets/app_icon.png" "$RESOURCES_DIR/AppIcon.png"
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

# 6. Sign and Notarize App Bundle (REQUIRED for macOS security)
ENTITLEMENTS="scripts/entitlements.plist"
if [ -n "$SIGNING_IDENTITY" ]; then
    echo "Signing binary and app bundle with identity: $SIGNING_IDENTITY"
    # Note: --options runtime is REQUIRED for notarization
    codesign --force --options runtime --entitlements "$ENTITLEMENTS" --sign "$SIGNING_IDENTITY" --timestamp "$MACOS_DIR/$BINARY_NAME"
    codesign --force --options runtime --entitlements "$ENTITLEMENTS" --sign "$SIGNING_IDENTITY" --timestamp "$APP_BUNDLE"

    if [ -n "$AC_APPLE_ID" ] && [ -n "$AC_PASSWORD" ] && [ -n "$AC_TEAM_ID" ]; then
        echo "Notarizing App Bundle..."
        ZIP_NAME="$BUILD_DIR/$APP_NAME-$ARCH.zip"
        rm -f "$ZIP_NAME"
        # Create a zip of the app bundle for notarization
        ditto -c -k --keepParent "$APP_BUNDLE" "$ZIP_NAME"
        
        xcrun notarytool submit "$ZIP_NAME" --apple-id "$AC_APPLE_ID" --password "$AC_PASSWORD" --team-id "$AC_TEAM_ID" --wait
        
        echo "Stapling App Bundle..."
        xcrun stapler staple "$APP_BUNDLE"
        rm -f "$ZIP_NAME"
        echo "✓ App Bundle notarization and stapling complete."
    else
        echo "⚠️ Notarization credentials missing. Skipping App notarization."
    fi
else
    echo "⚠️ SIGNING_IDENTITY not set. Skipping code signing. App will likely be blocked by Gatekeeper."
fi

# 7. Prepare DMG Root
# We need a folder that contains the .app AND a symlink to Applications
DMG_ROOT="$BUILD_DIR/dmg-root-$ARCH"
rm -rf "$DMG_ROOT"
mkdir -p "$DMG_ROOT"

# Copy the notarized .app bundle into the DMG root
cp -R "$APP_BUNDLE" "$DMG_ROOT/"

# Create symlink to Applications
ln -s /Applications "$DMG_ROOT/Applications"

# 8. Create DMG from the Root Folder
DMG_NAME="$BUILD_DIR/marubot-macos-$ARCH.dmg"
rm -f "$DMG_NAME"

echo "Creating DMG package..."
hdiutil create -volname "$APP_NAME v$VERSION" -srcfolder "$DMG_ROOT" -ov -format UDZO "$DMG_NAME"

# 9. Sign and Notarize DMG
if [ -n "$SIGNING_IDENTITY" ]; then
    echo "Signing DMG..."
    codesign --force --sign "$SIGNING_IDENTITY" --timestamp "$DMG_NAME"

    if [ -n "$AC_APPLE_ID" ] && [ -n "$AC_PASSWORD" ] && [ -n "$AC_TEAM_ID" ]; then
        echo "Submitting DMG for notarization..."
        xcrun notarytool submit "$DMG_NAME" --apple-id "$AC_APPLE_ID" --password "$AC_PASSWORD" --team-id "$AC_TEAM_ID" --wait
        
        echo "Stapling DMG..."
        xcrun stapler staple "$DMG_NAME"
        echo "✓ DMG notarization and stapling complete."
    else
        echo "⚠️ Notarization credentials missing. Skipping DMG notarization."
    fi
fi

# Cleanup
rm -rf "$DMG_ROOT"

echo "✓ Created $DMG_NAME"
if [ -n "$SIGNING_IDENTITY" ]; then
    echo "  (Signed and ready for distribution)"
fi
