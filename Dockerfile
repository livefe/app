FROM golang:1.24-alpine as builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /build

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod/

RUN go mod download

COPY . .

RUN go build -o /build/server cmd/server/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /build/server ./server
COPY --from=builder /build/config ./config/

EXPOSE 8080

CMD ["./server"]