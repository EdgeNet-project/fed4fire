dir_guard=@mkdir -p $(@D)

.PHONY: all
all: self_signed/client.pem self_signed/server.pem trusted_roots/ilabt.imec.be.pem verify

.PHONY: clean
clean:
	rm -rf self_signed/ trusted_roots/

.PHONY: verify
verify:
	openssl verify -CAfile self_signed/ca-client.pem self_signed/client.pem
	openssl verify -CAfile self_signed/ca-server.pem self_signed/server.pem

# Authority root certificate
self_signed/ca-client.pem:
	$(dir_guard)
	openssl req                        \
		-x509                          \
		-nodes                         \
		-newkey     rsa:4096           \
		-keyout     $(basename $@).key \
		-out        $@                 \
 		-days       365                \
 		-set_serial 0                  \
		-subj       '/CN=authority.localhost/emailAddress=ca-client@mail.com' \
		-addext     'subjectAltName = email:ca-client@mail.com, URI:urn:publicid:IDN+edge-net.org+authority+ca'

# Client certificate
# https://github.com/GENI-NSF/geni-docs/blob/master/GeniApiCertificates.adoc
self_signed/client.pem: self_signed/ca-client.pem
	$(dir_guard)
	echo 'subjectAltName = email:client@mail.com, URI:urn:publicid:IDN+edge-net.org+user+client, URI:urn:uuid:433b6339-43f0-4d88-b5f8-5709de6dff3b' > $(basename $@).ext
	openssl req                        \
		-nodes                         \
		-newkey     rsa:4096           \
		-keyout     $(basename $@).key \
		-out        $(basename $@).csr \
		-subj       '/CN=client'
	openssl x509                       \
		-req                           \
		-CA         $<                 \
		-CAkey      $(basename $<).key \
		-extfile    $(basename $@).ext \
		-in         $(basename $@).csr \
		-out        $@                 \
		-days       365                \
		-set_serial 0
	rm $(basename $@).csr $(basename $@).ext

# Server root certificate
self_signed/ca-server.pem:
	$(dir_guard)
	openssl req                        \
		-x509                          \
		-nodes                         \
		-newkey     rsa:4096           \
		-keyout     $(basename $@).key \
		-out        $@                 \
 		-days       365                \
 		-set_serial 0                  \
		-subj       '/CN=ca-server'

# Server certificate
self_signed/server.pem: self_signed/ca-server.pem
	$(dir_guard)
	openssl req                        \
		-nodes                         \
		-newkey     rsa:4096           \
		-keyout     $(basename $@).key \
		-out        $(basename $@).csr \
		-subj       '/CN=localhost'
	openssl x509                       \
		-req                           \
		-CA         $<                 \
		-CAkey      $(basename $<).key \
		-in         $(basename $@).csr \
		-out        $@                 \
		-days       365                \
		-set_serial 0
	rm $(basename $@).csr

trusted_roots/ilabt.imec.be.pem:
	$(dir_guard)
	curl -Lo trusted_roots/ilabt.imec.be.pem https://groups.geni.net/geni/raw-attachment/wiki/GeniTrustAnchors/ilabt.imec.be.pem
