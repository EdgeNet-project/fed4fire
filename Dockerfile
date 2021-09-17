FROM golang:1.16 as builder
WORKDIR /go/src/app
ADD . /go/src/app
RUN go get -d -v ./...
RUN go build -o /go/bin/app

FROM gcr.io/distroless/base
COPY --from=builder /go/bin/app /
ENTRYPOINT ["/app"]