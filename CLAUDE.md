# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CloudflareSpeedTest - A Go CLI tool that tests Cloudflare CDN IPs to find the fastest ones for better network performance. Supports both IPv4 and IPv6, with TCP and HTTP ping modes.

## Build Commands

```bash
# Standard build
go build

# Build with version string
go build -ldflags "-s -w -X main.version=v2.3.4"

# Cross-compilation (e.g., for Linux from macOS)
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=v2.3.4" -o cfst
```

## Architecture

The project follows a clear separation of concerns:

### Main Flow (`main.go`)
1. Parse command-line flags
2. Run latency tests via `task.NewPing().Run()`
3. Filter by delay/loss rate: `FilterDelay()` -> `FilterLossRate()`
4. Run download speed tests: `task.TestDownloadSpeed()`
5. Export CSV and print results

### Package Structure

**`task/`** - Core testing logic:
- `ip.go` - IP range parsing (CIDR notation), supports IPv4/IPv6, reads from file or `-ip` flag
- `tcping.go` - TCP ping latency testing with goroutine-based concurrency (max 1000)
- `httping.go` - HTTP ping alternative mode, extracts colo codes from CDN response headers
- `download.go` - Download speed testing, uses EWMA for speed calculation

**`utils/`** - Shared utilities:
- `csv.go` - CSV export and result sorting (implements `sort.Interface`)
- `color.go` - Colored terminal output (fatih/color)
- `progress.go` - Animated progress bars (cheggaaa/pb)

### Key Data Types

- `PingDelaySet` - IP results sorted by loss rate, then delay
- `DownloadSpeedSet` - IP results sorted by download speed (descending)
- `CloudflareIPData` - Contains IP, sent/received counts, delay, loss rate, download speed, colo code

### Colo Code Extraction (`task/httping.go:139-191`)

Supports extracting location codes from multiple CDN response headers:
- Cloudflare: `cf-ray` header (IATA 3-letter airport codes)
- AWS CloudFront: `x-amz-cf-pop` header
- Fastly: `x-served-by` header
- CDN77: `x-77-pop` header
- Bunny CDN: `server` header
- Gcore: `x-id-fe` header
