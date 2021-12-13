# Fed4FIRE Aggregate Manager for EdgeNet

[![Tests](https://img.shields.io/github/workflow/status/EdgeNet-project/fed4fire/Go?logo=github)](https://github.com/EdgeNet-project/fed4fire/actions/workflows/go.yml)
[![Coverage](https://img.shields.io/coveralls/github/EdgeNet-project/fed4fire?logo=coveralls&logoColor=white)](https://coveralls.io/github/EdgeNet-project/fed4fire)

## Development

```bash
make

go run main.go \
  -containerImage ubuntu2004:docker.io/library/ubuntu:20.04 \
  -kubeconfig ~/.kube/config \
  -parentNamespace lip6-lab-fed4fire-dev \
  -serverCert self_signed/server.pem \
  -serverKey self_signed/server.key \
  -trustedRootCert self_signed/ca-client.pem
```

## TODO
- Kube deployment
- Document why we need nginx and how it works
- Document why we need xmlsec1 and openssl bindings
- `env CGO_CFLAGS="-I/opt/homebrew/opt/openssl@1.1/include" CGO_LDFLAGS="-L/opt/homebrew/opt/openssl@1.1/lib" go get github.com/libp2p/go-openssl`

### Using jFed automated testing

#### Add local AM using jFed scanner

```bash
curl -LO https://jfed.ilabt.imec.be/downloads/stable/jar/jfed_gui.tar.gz
tar xvzf jfed_gui.tar.gz
cd jfed_gui/
# TODO: Download javafx (x86/arm64)
java --module-path ~/Downloads/javafx-sdk-17.0.0.1/lib/ -cp "*:lib/*" \
  --add-modules=javafx.swing,javafx.graphics,javafx.fxml,javafx.media,javafx.web \
  be.iminds.ilabt.jfed.ui.javafx.scanner.ScannerLauncher
```

TODO...

- The AM server is stateless, all the state is stored in the Kubernetes object through annotations.

## Mapping Fed4Fire concepts to Kubernetes

- Slice: namespace (here specifically EdgeNet subnamespaces)
- Sliver: deployment

Naming: first 8 bytes of a SHA512 hash in a hexadecimal string.
This allows to create objects with names that are valid in the GENI spec, but not in Kubernetes which mostly allows only alphanumeric chars.
