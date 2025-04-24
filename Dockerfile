# Stage 1: Build the Go application
FROM golang:1.23.0 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy && go mod download

COPY . .

# 👇 Compile for Alpine (static binary)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/server

# Stage 2: Minimal runtime image
FROM alpine:3.19.1

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=build /app/main .
COPY .env .

RUN chmod +x /app/main

ENTRYPOINT ["./main"]
CMD []
