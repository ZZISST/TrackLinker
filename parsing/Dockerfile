FROM golang:1.23.6

WORKDIR /app
COPY . .
RUN go mod download

RUN go build -o /app/parsing ./cmd

EXPOSE 8080
ENTRYPOINT ["/app/parsing"]