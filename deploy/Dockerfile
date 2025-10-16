# syntax=docker/dockerfile:1

# --- Builder stage ---
FROM golang:1.22-alpine AS builder


WORKDIR /injective-chronos-go

# Cache go modules
COPY go.mod go.sum ./
RUN go env -w GO111MODULE=on && \
 go env -w GOPROXY=https://goproxy.cn 

COPY . .

RUN go mod download
# Build (honor buildx provided TARGETOS/TARGETARCH if present)
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags "-s -w" -o /out/chronos-api ./cmd

# --- Runtime stage ---
FROM alpine:3.19


WORKDIR /app

# Binary
COPY --from=builder /out/chronos-api /app/chronos-api

# Config (default path used by the app)
COPY etc /app/etc

# Writable log directory for non-root runtime
RUN mkdir -p /app/log && chown -R 65532:65532 /app

EXPOSE 4442

# 这行用于以非 root 用户（UID=65532, GID=65532）运行容器进程，提高安全性
USER 65532:65532

ENTRYPOINT ["/app/chronos-api"]


