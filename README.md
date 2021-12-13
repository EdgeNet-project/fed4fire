# Fed4FIRE Aggregate Manager for EdgeNet

[![CI/Build](https://img.shields.io/github/workflow/status/EdgeNet-project/fed4fire/Go?logo=github&label=build)](https://github.com/EdgeNet-project/fed4fire/actions/workflows/go.yml)
[![CI/Docker](https://img.shields.io/github/workflow/status/EdgeNet-project/fed4fire/Docker?logo=github&label=docker)](https://github.com/EdgeNet-project/fed4fire/actions/workflows/docker.yml)
[![Coverage](https://img.shields.io/coveralls/github/EdgeNet-project/fed4fire?logo=coveralls&logoColor=white)](https://coveralls.io/github/EdgeNet-project/fed4fire)

This package implements the [GENI Aggregate Manager API Version 3](https://groups.geni.net/geni/wiki/GAPI_AM_API_V3) to federate [EdgeNet](https://www.edge-net.org) under the [Fed4FIRE+](https://www.fed4fire.eu) project.

## Accessing EdgeNet through Fed4FIRE

- To run experiments on a Fed4FIRE testbed, follow the instructions at https://doc.fed4fire.eu
- EdgeNet defines one non-exclusive sliver type named `container` and one disk image named `ubuntu2004`
- The SSH key must be set through the `geni_update_users` operational action

## Architecture

- The AM server is stateless, all the information about slices and slivers is stored in Kubernetes objects annotations
- Slices maps to Kubernetes namespaces
- Slivers maps to Kubernetes deployments
- Object names are derived from the first 8 bytes of the SHA512 hash of the RSpec name. This allows to create objects with names that are valid in the GENI spec, but not in Kubernetes which mostly allows only alphanumeric chars.

### Workarounds

- Fed4FIRE uses client certificates with non-standard OIDs that are not supported by the Go X.509 parser. As such we rely on nginx to verify the client certificate and pass the decoded certificate to the AM server. The openssl CLI tool is then used to process the certificate, instead of the Go standard library.

## Deployment

The AM image is hosted on Docker Hub ([`edgenetio/fed4fire`](https://hub.docker.com/r/edgenetio/fed4fire)):
```bash
docker run edgenetio/fed4fire:main --help
```

The AM must be deployed behind a reverse proxy that pass the `X-Fed4Fire-Certificate` header.
For an example, see [`dev/nginx.conf`](https://github.com/EdgeNet-project/fed4fire/blob/main/dev/nginx.conf).

## Development

```bash
git clone git@github.com:EdgeNet-project/fed4fire.git && cd fed4fire/dev/
# Create a self-signed server certificate and download the trusted client root certificates
make
# Start the AM behind nginx
docker-compose up
# Optionnally, connect to the Go debug server
dlv connect localhost:40000
# Issue XML-RPC calls (set `--cert` to the appropriate client certificate path)
curl --cacert self_signed/ca-server.pem \
     --cert ~/.jFed/login-certs/*.pem \
     --data '<methodCall><methodName>GetVersion</methodName><params/></methodCall>' \
     --header "Content-Type: text/xml" \
     --request POST \
     https://localhost:9443
```
