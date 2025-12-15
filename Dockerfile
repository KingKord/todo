FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o /bin/server ./cmd/server

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /bin/server /bin/server
EXPOSE 50051

ENV GRPC_ADDR=:50051
ENV POSTGRES_DSN=postgres://postgres:postgres@db:5432/todos?sslmode=disable

CMD ["/bin/server"]
