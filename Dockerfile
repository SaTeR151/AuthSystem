FROM golang:1.24.4

WORKDIR /app

COPY . .

RUN go mod tidy