# WebServer

**Source Spec:** main.md

## Responsibilities

### Knows
- contentRoot: Directory containing HTML files or BundleManager for bundled mode
- port: HTTP server port (same as WebSocket port)
- indexHTML: Cached index.html content for SPA routing

### Does
- serveFile: Serve static file from content root
- handleSPARoute: Detect route (no extension) and serve index.html while preserving URL
- detectFileType: Determine content type from file extension
- return404: Return 404 for missing files with extensions

## Collaborators

- BundleManager: Reads files from bundled content in bundled mode
- Server: Started/stopped by Server

## Sequences

- seq-http-serve-file.md: Static file serving
- seq-spa-routing.md: SPA route fallback to index.html
