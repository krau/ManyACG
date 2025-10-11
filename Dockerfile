FROM golang:alpine AS builder
WORKDIR /app

RUN apk add --no-cache git bash

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG BUILT_AT
ARG GIT_COMMIT
ARG VERSION

RUN builtAt=${BUILT_AT:-$(date +'%F %T %z')} && \
    gitCommit=${GIT_COMMIT:-$(git log --pretty=format:"%h" -1)} && \
    version=${VERSION:-$(git describe --abbrev=0 --tags)} && \
    ldflags="\
    -w -s \
    -X 'github.com/krau/ManyACG/internal/common/version/version.BuildTime=$builtAt' \
    -X 'github.com/krau/ManyACG/internal/common/version/version.Commit=$gitCommit' \
    -X 'github.com/krau/ManyACG/internal/common/version/version.Version=$version'\
    " && \
    CGO_ENABLED=0 go build -tags without_vips,nodynamic -ldflags "$ldflags" -o manyacg

FROM alpine:latest
WORKDIR /opt/manyacg/

RUN apk add --no-cache bash ca-certificates ffmpeg && update-ca-certificates

COPY --from=builder /app/manyacg .

ENTRYPOINT ["./manyacg"]
