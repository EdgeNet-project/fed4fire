FROM golang:1.17 as builder
ARG GCFLAGS=""
WORKDIR /go/src/app

RUN go get github.com/go-delve/delve/cmd/dlv

ADD go.mod go.mod
ADD go.sum go.sum
RUN go mod download

ADD . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -gcflags="$GCFLAGS" -o /go/bin/app

FROM debian:11

RUN apt-get update \
    && apt-get install --no-install-recommends --yes \
      openssl xmlsec1 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /go/bin/app /
COPY --from=builder /go/bin/dlv /

ENTRYPOINT ["/app"]
