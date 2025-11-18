# BundleManager

**Source Spec:** main.md

## Responsibilities

### Knows
- zipOffset: Offset of ZIP archive in binary
- zipSize: Size of ZIP archive
- zipReader: ZIP archive reader
- footer: Magic marker + offset + size (24 bytes total)

### Does
- checkBundled: Check if binary has bundled content (magic marker present)
- readFile: Read file from bundled ZIP archive
- listFiles: List all files in bundled archive
- copyFiles: Copy files matching glob patterns from bundle to destination
- extractAll: Extract entire bundle to current directory
- appendBundle: Append ZIP + footer to binary to create bundled executable

## Collaborators

- WebServer: Provides files when serving in bundled mode
- ExtractCommand: Extracts bundle to filesystem
- BundleCommand: Creates new bundled binary
- LsCommand: Lists bundled files
- CpCommand: Copies files from bundle

## Sequences

- seq-bundle-check.md: Detecting bundled content at startup
- seq-bundle-create.md: Creating bundled executable
- seq-bundle-extract.md: Extracting bundle to directory
- seq-bundle-read.md: Reading files from bundle
