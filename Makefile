dir_guard=@mkdir -p $(@D)

.PHONY: all
all: self_signed/server.pem trusted_roots/ilabt.imec.be.pem verify

.PHONY: clean
clean:
	rm -rf self_signed/ trusted_roots/

.PHONY: verify
verify:
	openssl verify -CAfile self_signed/ca-server.pem self_signed/server.pem

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
