# Run this script using `ruby package.rb` with an executable named `app`
# in a `build/` directory.

# Build an app bundle for macOS
def build_macos_package

  require 'fileutils'

  info_plist = %(
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleExecutable</key>
  <string>mcd</string>
  <key>CFBundleIconFile</key>
  <string>app.icns</string>
  <key>CFBundleInfoDictionaryVersion</key>
  <string>6.0</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>CFBundleVersion</key>
  <string>1</string>
  <key>NSHighResolutionCapable</key>
  <string>True</string>
</dict>
</plist>
)

  # Create directories
  FileUtils.mkpath 'pkgbuilddir/mcd.app/Contents/MacOS'
  FileUtils.mkpath 'pkgbuilddir/mcd.app/Contents/Resources'

  # Create Info.plist and copy over assets
  File.open('pkgbuilddir/mcd.app/Contents/Info.plist', 'w') { |f| f.write(info_plist) }
  FileUtils.cp 'buildd/mcd', 'pkgbuilddir/mcd.app/Contents/MacOS/'

  # Success
  puts 'App written to `pkgbuilddir/mcd.app`.'
end

build_macos_package