.PHONY: all
all: client server

.PHONY: clean
clean:
	rm -f self_signed/*.{csr,key,pem}

.PHONY: client
client:
	mkdir -p self_signed/
	# Authority certificate
	openssl req -nodes -x509 -newkey rsa:4096 -keyout self_signed/ca-client.key -out self_signed/ca-client.pem -days 365 \
		-subj '/CN=authority.localhost/emailAddress=ca-client@mail.com' -set_serial 0 \
		-addext "subjectAltName = email:ca-client@mail.com, URI:urn:publicid:IDN+edge-net.org+authority+ca"
	# Client certificate
	# https://github.com/GENI-NSF/geni-docs/blob/master/GeniApiCertificates.adoc
	openssl req -nodes -newkey rsa:4096 -keyout self_signed/client.key -out self_signed/client.csr -subj '/CN=client'
	echo "subjectAltName = email:client@mail.com, URI:urn:publicid:IDN+edge-net.org+user+client, URI:urn:uuid:433b6339-43f0-4d88-b5f8-5709de6dff3b" > self_signed/client.ext
	openssl x509 -req -CA self_signed/ca-client.pem -CAkey self_signed/ca-client.key -in self_signed/client.csr -out self_signed/client.pem -extfile self_signed/client.ext -set_serial 0 -days 365
	rm self_signed/client.csr self_signed/client.ext
	# Verify client certificate
	openssl verify -CAfile self_signed/ca-client.pem self_signed/client.pem

.PHONY: server
server:
	mkdir -p self_signed/
	# Authority certificate
	openssl req -nodes -x509 -newkey rsa:4096 -keyout self_signed/ca-server.key -out self_signed/ca-server.pem -days 365 -subj '/CN=server-authority'
	# Server certificate
	openssl req -nodes -newkey rsa:4096 -keyout self_signed/server.key -out self_signed/server.csr -days 365 -subj '/CN=localhost'
	openssl x509 -req -CA self_signed/ca-server.pem -CAkey self_signed/ca-server.key -in self_signed/server.csr -out self_signed/server.pem -set_serial 0 -days 365
	rm self_signed/server.csr
	# Verify server certificate
	openssl verify -CAfile self_signed/ca-server.pem self_signed/server.pem
