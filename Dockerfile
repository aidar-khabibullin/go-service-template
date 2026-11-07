ARG GO_VERSION=1.26

# Modules caching
FROM golang:${GO_VERSION}-alpine AS modules

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# Build
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

COPY --from=modules /go/pkg /go/pkg
COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /my-app ./cmd/app

# Runtime: minimal image, no shell, non-root user
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /my-app /my-app

USER nonroot:nonroot

ENTRYPOINT ["/my-app"]
