.PHONY: all
all: client server

.PHONY: clean
clean:
	rm -f *.{csr,key,pem}

.PHONY: client
client:
	# Authority certificate
	openssl req -nodes -x509 -newkey rsa:4096 -keyout ca-client.key -out ca-client.pem -days 365 -subj '/CN=client-authority'
	# Client certificate
	openssl req -nodes -newkey rsa:4096 -keyout client.key -out client.csr -days 365 -subj '/CN=client'
	openssl x509 -req -CA ca-client.pem -CAkey ca-client.key -in client.csr -out client.pem -set_serial 01 -days 365
	rm client.csr
	# Verify client certificate
	openssl verify -CAfile ca-client.pem client.pem

.PHONY: server
server:
	# Authority certificate
	openssl req -nodes -x509 -newkey rsa:4096 -keyout ca-server.key -out ca-server.pem -days 365 -subj '/CN=server-authority'
	# Server certificate
	openssl req -nodes -newkey rsa:4096 -keyout server.key -out server.csr -days 365 -subj '/CN=localhost'
	openssl x509 -req -CA ca-server.pem -CAkey ca-server.key -in server.csr -out server.pem -set_serial 01 -days 365
	rm server.csr
	# Verify server certificate
	openssl verify -CAfile ca-server.pem server.pem
