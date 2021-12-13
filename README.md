# Fed4FIRE Aggregate Manager for EdgeNet

[![CI/Build](https://img.shields.io/github/workflow/status/EdgeNet-project/fed4fire/Go?logo=github&label=build)](https://github.com/EdgeNet-project/fed4fire/actions/workflows/go.yml)
[![CI/Docker](https://img.shields.io/github/workflow/status/EdgeNet-project/fed4fire/Docker?logo=github&label=docker)](https://github.com/EdgeNet-project/fed4fire/actions/workflows/docker.yml)
[![Coverage](https://img.shields.io/coveralls/github/EdgeNet-project/fed4fire?logo=coveralls&logoColor=white)](https://coveralls.io/github/EdgeNet-project/fed4fire)

This package implements the [GENI Aggregate Manager API Version 3](https://groups.geni.net/geni/wiki/GAPI_AM_API_V3) to federate [EdgeNet](https://www.edge-net.org) under the [Fed4FIRE+](https://www.fed4fire.eu) project.

## Accessing EdgeNet through Fed4FIRE

## Architecture

- The AM server is stateless, all the state is stored in the Kubernetes object through annotations.

### Mapping Fed4Fire concepts to Kubernetes

- Slice: namespace (here specifically EdgeNet subnamespaces)
- Sliver: deployment

Naming: first 8 bytes of a SHA512 hash in a hexadecimal string.
This allows to create objects with names that are valid in the GENI spec, but not in Kubernetes which mostly allows only alphanumeric chars.

### Workarounds

## Deployment

The AM must be deployed behind a reverse proxy that pass the `X-Fed4Fire-Certificate` header.
For an example, see [`dev/nginx.conf`](https://github.com/EdgeNet-project/fed4fire/blob/main/dev/nginx.conf).

To see the AM options:
```bash
docker run edgenetio/fed4fire:main --help
```

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
