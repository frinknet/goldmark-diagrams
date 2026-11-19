# syntax=docker/dockerfile:1

FROM caddy:2.11.9-builder-alpine AS build

RUN apk add --no-cache ca-certificates tzdata curl upx

WORKDIR /src

# cache the module downloads separately from the source
COPY go.mod go.sum ./
RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
    go mod download

COPY . .

# pick up any newly uncommented imports
RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
    go mod tidy

RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /bin/goldmark-diagrams ./cmd/goldmark-diagrams

ENTRYPOINT ["/bin/goldmark-diagrams"]
CMD [""]
