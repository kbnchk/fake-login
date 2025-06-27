FROM golang:alpine AS builder
WORKDIR /build
ADD go.mod .
ADD main.go .
RUN go build -o fake-login main.go
FROM scratch
WORKDIR /app
COPY --from=builder /build/fake-login .
COPY static/ ./static/
ENTRYPOINT [ "/app/fake-login" ]