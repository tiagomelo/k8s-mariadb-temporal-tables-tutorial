FROM golang:alpine AS builder

RUN apk update && apk add bash

# Move to working directory /app
WORKDIR /app

# Copy the code into the container
COPY . .

# Download dependencies using go mod
RUN go mod download

# Build the application's binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o cmd/main cmd/main.go

# Build a smaller image that will only contain the application's binary
FROM alpine:3.11.3

RUN apk update && apk add bash

# Move to working directory /app
WORKDIR /app

# Copy application's binary
COPY --from=builder /app .

# Command to run the application when starting the container
CMD ["./cmd/main"]
