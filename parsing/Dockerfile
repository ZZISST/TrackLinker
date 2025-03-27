FROM golang:1.23.6

WORKDIR /app
COPY . .
RUN go mod download

COPY ./pkg/config/config.sample.yaml ./pkg/config/config.yaml
RUN go build -o /app/parsing ./cmd

EXPOSE 8888
ENTRYPOINT ["/app/parsing"]