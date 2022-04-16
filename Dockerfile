FROM golang:1.18 AS multi-stage-build

ENV GOOS linux
ENV GOARCH amd64

RUN mkdir -p /go/src/github.com/newtstat/cloudacme
WORKDIR /go/src/github.com/newtstat/cloudacme
COPY go.mod /go/src/github.com/newtstat/cloudacme
COPY go.sum /go/src/github.com/newtstat/cloudacme
RUN go mod download

COPY . /go/src/github.com/newtstat/cloudacme

# NOTE: REVISION and TIMESTAMP change each time, so they must be placed after go mod download for docker build cache.
ARG VERSION
ARG REVISION
ARG BRANCH
ARG TIMESTAMP

RUN go build -ldflags "-X github.com/newtstat/cloudacme/config.version=${VERSION} -X github.com/newtstat/cloudacme/config.revision=${REVISION} -X github.com/newtstat/cloudacme/config.branch=${BRANCH} -X github.com/newtstat/cloudacme/config.timestamp=${TIMESTAMP}" ./cmd/cloudacme/...

FROM debian:11-slim
# NOTE: Best practices for writing Dockerfiles | Docker Documentation https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#run
RUN useradd --home-dir /app --create-home app && \
  apt-get update -qqy && \
  apt-get install -qqy --no-install-recommends ca-certificates curl dumb-init iproute2 net-tools && \
  rm -rf /var/lib/apt/lists/*
USER app
WORKDIR /app
COPY --from=multi-stage-build --chown=app:app /go/src/github.com/newtstat/cloudacme/cloudacme /app/cloudacme
CMD ["dumb-init", "--", "/app/cloudacme"]
