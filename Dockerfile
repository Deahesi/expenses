FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -tags timetzdata \
    -ldflags="-s -w" \
    -o /main ./cmd/server


FROM builder AS migrate

RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.26.0

ENTRYPOINT ["/go/bin/goose"]


FROM gcr.io/distroless/static-debian12:latest AS runtime

COPY --from=builder /main /main

EXPOSE 8080

ENTRYPOINT ["/main"]