FROM golang:1.24.3-alpine AS build  

WORKDIR /app
COPY go.mod .
COPY main.go .

RUN go build -o app main.go

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/app .

CMD ["./app"]
