# Fed4FIRE Aggregate Manager for EdgeNet

[![Go](https://github.com/EdgeNet-project/fed4fire/actions/workflows/go.yml/badge.svg)](https://github.com/EdgeNet-project/fed4fire/actions/workflows/go.yml)

❗This repository is currently private during the design phase. We'll squash the commits and make it public at a later time.

TODO: Add coverage/docker badges.

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

## Mapping Fed4Fire concepts to Kubernetes

- Slice: namespace (here specifically EdgeNet subnamespaces)
- Sliver: deployment

Naming: first 8 bytes of a SHA512 hash in a hexadecimal string.
This allows to create objects with names that are valid in the GENI spec, but not in Kubernetes which mostly allows only alphanumeric chars.

## [AM API requirements](https://doc.fed4fire.eu/testbed_owner/addingtestbed.html#requirements)

- A server to run the AM software on.
  - We'll run it on EdgeNet as a pod.
- A publicly reachable IP for that server. This needs to be either an IPv4 or an IPv6 address. We recommend both.
  - Ok.
- A DNS name for that server, that resolves to the publicly reachable IP addresses of the server. (Recommendation: It’s nice if the DNS name refers to your testbed and is specific for your AM. Example: `am.mytestbed.example.com`)
  - `am.edge-net.org`?
  - Full AM URL: `https://am.edge-net.org/protogeni/xmlrpc/am/3.0`?
- [Choose a URN for your AM](https://doc.fed4fire.eu/testbed_owner/rspec.html#choosing-your-component-manager-urn). This is of the form: `urn:publicid:IDN+DNSNAME+authority+am` where your replace DNSNAME by the DNS name of your AM. (Example: `urn:publicid:IDN+mytestbed.example.com+authority+am`).
  - `urn:publicid:IDN+edge-net.org+authority+am`
- Choose a port at which you server will run. There is no standard port in the specification, so a lot of different ports are used in practice (12369, 8010, …). We recommend using port 443, if that is not in use by anything else. The advantage of using the default https port is that it is reachable through most firewalls, and the protocol is in fact using https.
  - `443`?
- You need a X.509 Server Certificate, because the AM server uses https. This can be a self signed certificate (jFed stores a list of these to make it work safely). However, in that case, make sure you configure the fields in your self signed certificate correctly. [See the next section for more details](https://doc.fed4fire.eu/testbed_owner/addingtestbed.html#server-x-509-certificate).
  - “Subject” field of the certificate must contain a “CN” that is the hostname of the server (NOT and IP, the DNS hostname!)
    - `CN=am.edge-net.org`
  - The “X509v3 Subject Alternative Name” section, must contain a “DNS” entry, which is the hostname of your server (NOT and IP, the DNS hostname!). Note that this means that your AM needs a DNS name, not just an IP address!
    - `DNS:am.edge-net.org`
  - In production we can probably use a LetsEncrypt certificate provide by `cert-manager` in the cluster (@Berat is this possible?).
- You probably have testbed resources that you want to make reachable to experimenters using SSH. There are 2 options: You need public IP addresses that you can assign to these nodes when needed (IPv4 or IPv6). You need to have one machine with a publically reachable IP address (IPv4 recommended) act as a gateway.
  - We can assign public IP, though we will use multiple SSH ports (different from `22`) to allow for multiple "resources" on the same node.
