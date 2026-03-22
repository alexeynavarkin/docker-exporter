FROM golang:1.25-alpine AS builder

WORKDIR /build
COPY . .
RUN go build -o docker-exporter ./cmd/main.go

FROM alpine:3.18
WORKDIR /docker-exporter
COPY --from=builder /build/docker-exporter .
CMD [ "./docker-exporter" ]