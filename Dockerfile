FROM golang:1.25-alpine AS dev

WORKDIR /app

RUN apk add --no-cache git ca-certificates

RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["air", "-c", ".air.toml"]


FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /auto_park ./cmd/auto_park


FROM alpine:3.20 AS production

WORKDIR /app

COPY --from=builder /auto_park /usr/local/bin/auto_park

EXPOSE 8080

CMD ["/usr/local/bin/auto_park"]