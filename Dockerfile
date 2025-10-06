FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .

ENV CGO_ENABLED=0
RUN go get && go build -o mm

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/mm .

CMD ["./mm"]
