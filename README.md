# fake-login

A plausible login page ("Northgate Portal") that answers any request without a route of its own:
access by IP, unknown SNI/Host. It works as a decoy so the server doesn't look "empty" to scanners.

- Single static binary (`scratch` image, user 65534); static files are embedded via `go:embed`, the page makes no external requests.
- `POST /api/login` always responds with 401 after 0.7–1.4 s; credentials are never stored or logged. At most 64 requests are handled concurrently, the body is limited to 4 KB.
- Pages: `/`, `/forgot-password`, `/privacy`, `/terms`, a custom `404`, `robots.txt`, `favicon`.
- Security headers and a CSP without inline code, ETag, server timeouts (Slowloris), graceful shutdown.
- `fake-login -healthcheck` for Docker `HEALTHCHECK`.

## Environment variables
| Variable | Default | Description |
|---|---|---|
| `SERVER_ADDR` | `:8080` | listen address |
| `TLS_CERT_FILE`, `TLS_KEY_FILE` | empty | TLS is enabled if both files exist |
| `TLS_MIN_VER` | `1.3` | minimum TLS version |

## Build and release
`docker build -t fake-login .`; the image is built and pushed to GHCR by GitHub Actions when a `v*`/`x.y.z` tag is pushed.
