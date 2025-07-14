# --- Build stage ---
FROM golang:1.24.3-alpine AS build

WORKDIR /app
COPY go.mod .
COPY main.go .

RUN go build -o app main.go

# --- Run stage: distroless & non-root (rootless) ---
FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=build /app/app .
COPY myimage.jpg .

# Run as non-root user (UID 1000 always exists in distroless)
USER 1000:1000

CMD ["./app"]
