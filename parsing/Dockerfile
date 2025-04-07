FROM golang:1.23.6

WORKDIR /app

# Copy go.mod and go.sum first to cache dependency installation
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
RUN go build -o /app/parsing ./cmd

EXPOSE 8080
ENTRYPOINT ["/app/parsing"]