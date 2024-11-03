FROM golang:1.22 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY ./app /app
RUN go build -o /app/main .

RUN apt-get update && apt-get install -y librsvg2-bin webp && rm -rf /var/lib/apt/lists/*

CMD ["/app/main"]