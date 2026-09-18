# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/syspeek ./cmd/syspeek

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/syspeek /usr/local/bin/syspeek

ENTRYPOINT ["/usr/local/bin/syspeek"]
CMD [""]
