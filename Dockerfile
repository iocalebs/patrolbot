FROM golang:1.25 AS build
WORKDIR /app
COPY go.mod go.sum .
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
        -mod=readonly \
        -trimpath \
        -ldflags="-s -w" \
        -o patrolbot

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build --chown=nonroot:nonroot /app/patrolbot /patrolbot
COPY config.yaml /config.yaml
ENTRYPOINT ["/patrolbot"]
CMD ["serve", "--config", "/config.yaml"]
