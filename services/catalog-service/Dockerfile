FROM golang:1.23-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./services/catalog-service/cmd/server
RUN CGO_ENABLED=0 GOBIN=/out go install github.com/grpc-ecosystem/grpc-health-probe@v0.4.25

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/server /app/server
COPY --from=builder /out/grpc-health-probe /usr/local/bin/grpc-health-probe
COPY services/catalog-service/migrations /app/migrations
USER nonroot:nonroot
EXPOSE 50053
ENTRYPOINT ["/app/server"]
