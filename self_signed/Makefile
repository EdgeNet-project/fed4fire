.PHONY: all
all: client server

.PHONY: clean
clean:
	rm -f *.{csr,key,pem}

.PHONY: client
client:
	# Authority certificate
	openssl req -nodes -x509 -newkey rsa:4096 -keyout ca-client.key -out ca-client.pem -days 365 \
		-subj '/CN=authority.localhost/emailAddress=ca-client@mail.com' -set_serial 0 \
		-addext "subjectAltName = email:ca-client@mail.com, URI:urn:publicid:IDN+edge-net.org+authority+ca"
	# Client certificate
	# https://github.com/GENI-NSF/geni-docs/blob/master/GeniApiCertificates.adoc
	openssl req -nodes -newkey rsa:4096 -keyout client.key -out client.csr -subj '/CN=client'
	echo "subjectAltName = email:client@mail.com, URI:urn:publicid:IDN+edge-net.org+user+client, URI:urn:uuid:433b6339-43f0-4d88-b5f8-5709de6dff3b" > client.ext
	openssl x509 -req -CA ca-client.pem -CAkey ca-client.key -in client.csr -out client.pem -extfile client.ext -set_serial 0 -days 365
	rm client.csr client.ext
	# Verify client certificate
	openssl verify -CAfile ca-client.pem client.pem

.PHONY: server
server:
	# Authority certificate
	openssl req -nodes -x509 -newkey rsa:4096 -keyout ca-server.key -out ca-server.pem -days 365 -subj '/CN=server-authority'
	# Server certificate
	openssl req -nodes -newkey rsa:4096 -keyout server.key -out server.csr -days 365 -subj '/CN=localhost'
	openssl x509 -req -CA ca-server.pem -CAkey ca-server.key -in server.csr -out server.pem -set_serial 0 -days 365
	rm server.csr
	# Verify server certificate
	openssl verify -CAfile ca-server.pem server.pem
