FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY go.mod .
COPY main.go .
COPY static/ ./static/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o fake-login .

FROM scratch
WORKDIR /app
COPY --from=builder /build/fake-login .
USER 65534:65534
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 CMD ["/app/fake-login", "-healthcheck"]
ENTRYPOINT ["/app/fake-login"]
