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
      ca-certificates curl openssl unzip xmlsec1 \
    && rm -rf /var/lib/apt/lists/*

# Add kubectl and kubelogin for development with local OIDC kubeconfig file.
RUN curl -L "https://dl.k8s.io/release/v1.21.5/bin/linux/$(arch | sed -e "s/aarch64/arm64/")/kubectl" > /usr/bin/kubectl && \
    chmod +x /usr/bin/kubectl
RUN curl -L "https://github.com/int128/kubelogin/releases/download/v1.25.1/kubelogin_linux_$(arch | sed -e "s/aarch64/arm64/").zip" > kubelogin.zip && \
    unzip kubelogin.zip && \
    rm kubelogin.zip && \
    mv kubelogin /usr/bin/kubectl-oidc_login

COPY --from=builder /go/bin/app /
COPY --from=builder /go/bin/dlv /

ENTRYPOINT ["/app"]
