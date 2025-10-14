# Echo Server

A versatile echo server with JA3 fingerprinting, file upload support, and HMAC authentication.

## Features

- 🔐 **JA3 Fingerprinting** - TLS client fingerprinting
- 📤 **File Upload** - Secure file upload with HMAC authentication
  - Path traversal attack prevention
  - Automatic path sanitization
  - Directory boundary validation
- 🛡️ **WAF Middleware** - Web application firewall
- 🔑 **HMAC Auth** - Flexible authentication middleware
- 🌐 **WebSocket Support** - Echo WebSocket server
- 📍 **IP Geolocation** - MaxMind GeoIP integration (GeoLite2)

## Quick Start

### Installation

```bash
go build -o echo-server .
```

### Usage

View available options:

```bash
./echo-server --help
```

Output:
```
Usage: echo-server [flags]

A versatile echo server with JA3 fingerprinting and file upload support

Flags:
  -h, --help                    Show context-sensitive help.
      --http="0.0.0.0:9012"     HTTP bind address ($HTTP_ADDR)
      --https="0.0.0.0:9013"    HTTPS bind address ($HTTPS_ADDR)
      --cert=STRING             TLS certificate file path ($TLS_CERT_FILE)
      --key=STRING              TLS key file path ($TLS_KEY_FILE)
      --debug                   Enable debug logging
      --auth-paths=/upload,/upload/*,...
                                Paths requiring HMAC authentication (supports
                                wildcards) ($AUTH_PATHS)
```

### Examples

**Start HTTP server:**
```bash
./echo-server --http 0.0.0.0:8080
```

**Start HTTPS server:**
```bash
./echo-server --cert cert.pem --key key.pem --https 0.0.0.0:9443
```

**Start both HTTP and HTTPS:**
```bash
./echo-server \
  --http 0.0.0.0:8080 \
  --https 0.0.0.0:8443 \
  --cert cert.pem \
  --key key.pem
```

**Using environment variables:**
```bash
export HTTP_ADDR=0.0.0.0:8080
export TLS_CERT_FILE=/path/to/cert.pem
export TLS_KEY_FILE=/path/to/key.pem
./echo-server
```

## Testing

### Echo endpoint:
```bash
curl -v https://echo.skiloop.com/abc
```

### Upload file:
```bash
# See examples/test_upload.sh for upload examples
cd examples
./test_upload.sh
```

## Documentation

- [File Upload API](docs/UPLOAD_API.md) - Complete upload API documentation
- [Upload Quick Start](docs/UPLOAD_QUICKSTART.md) - Quick start guide
- [Upload Security](docs/UPLOAD_SECURITY.md) - Path security and attack prevention ⭐
- [Auth Middleware](middleware/AUTH_MIDDLEWARE.md) - HMAC authentication middleware
- [Auth Paths Configuration](docs/AUTH_PATHS_CONFIG.md) - Configure authentication paths
- [Kong CLI Refactoring](docs/KONG_CLI_REFACTORING.md) - Command-line parameter documentation

## Configuration

### Command-line Parameters

| Parameter | Description | Default | Environment Variable |
|-----------|-------------|---------|---------------------|
| `--http` | HTTP bind address | `0.0.0.0:9012` | `HTTP_ADDR` |
| `--https` | HTTPS bind address | `0.0.0.0:9013` | `HTTPS_ADDR` |
| `--cert` | TLS certificate file | - | `TLS_CERT_FILE` |
| `--key` | TLS key file | - | `TLS_KEY_FILE` |
| `--debug` | Enable debug logging | `true` | - |
| `--auth-paths` | Paths requiring HMAC auth | `/upload,/upload/*` | `AUTH_PATHS` |

### Environment Variables

**Server:**
- `HTTP_ADDR` - HTTP bind address
- `HTTPS_ADDR` - HTTPS bind address
- `TLS_CERT_FILE` - TLS certificate file path
- `TLS_KEY_FILE` - TLS key file path

**Authentication:**
- `AUTH_API_KEY` - API key for HMAC authentication
- `ECHO_AUTH_TIMESTAMP_VALID` - Timestamp validity in seconds (default: 300)
- `AUTH_PATHS` - Comma-separated paths requiring authentication (default: /upload,/upload/*)

**Upload:**
- `UPLOAD_DIR` - Upload directory (default: ./uploads)
- `UPLOAD_MAX_SIZE` - Max file size in bytes (default: 10485760)

**IP Geolocation:**
- `GEO_LITE_2_PATH` - Path to GeoLite2 database

## API Endpoints

### Echo Endpoints

- `GET/POST/PUT/PATCH /echo/*` - Echo request
- `GET/POST/PUT/PATCH /json/*` - Echo as JSON

### File Upload (requires HMAC auth)

- `POST /upload` - Upload file

### Utilities

- `GET /ok` - Health check
- `GET /health` - Health check
- `GET /_health` - Health check
- `GET /ws/echo` - WebSocket echo

### JA3 Fingerprinting

- `GET /ja3` - Get JA3 fingerprint

### IP Geolocation

Set `GEO_LITE_2_PATH` environment variable to enable IP location APIs.

## License

MIT
