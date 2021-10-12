FROM golang:1.17.1 as builder
ARG GCFLAGS=""
WORKDIR /go/src/app

RUN go get github.com/go-delve/delve/cmd/dlv

ADD go.mod go.mod
ADD go.sum go.sum
RUN go mod download

ADD . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -gcflags="$GCFLAGS" -o /go/bin/app

FROM gcr.io/distroless/base
COPY --from=builder /go/bin/app /
COPY --from=builder /go/bin/dlv /
ENTRYPOINT ["/app"]
