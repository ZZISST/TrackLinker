FROM golang:1.23.6

WORKDIR /app

# Copy go.mod and go.sum first to cache dependency installation
COPY . . 
RUN go mod download

# Build the application
RUN go build -o /app/parsing ./cmd

EXPOSE 8080
ENTRYPOINT ["/app/parsing"]