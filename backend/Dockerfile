FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/api ./cmd/api

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /app/api /app/api
EXPOSE 8080
ENTRYPOINT ["/app/api"]
