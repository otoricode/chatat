FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o chatat-server cmd/server/main.go

FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/chatat-server .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./chatat-server"]
