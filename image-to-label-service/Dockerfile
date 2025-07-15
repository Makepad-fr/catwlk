# --- Build stage ---
FROM golang:1.24.3-bookworm AS build

WORKDIR /app
COPY go.mod .
COPY main.go .
ENV CGO_ENABLED=0

RUN go build -o app main.go

# --- Run stage: distroless & non-root (rootless) ---
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
USER nonroot:nonroot

COPY --from=build --chown=nonroot:nonroot /app/app .
COPY ./public/index.html ./public/       


# Run as non-root user (UID 1000 always exists in distroless)


ENTRYPOINT ["/app/app"]
