PREFIX := $(GOPATH)
BINDIR := $(PREFIX)/bin
SOURCE := *.go router/*.go middleware/*.go middleware/*/*.go backend/*.go backend/*/*.go

all: fmt mohawk

mohawk: $(SOURCE)
	go build -o mohawk *.go

.PHONY: fmt
fmt: $(SOURCE)
	gofmt -s -l -w $(SOURCE)

.PHONY: clean
clean:
	$(RM) mohawk

.PHONY: test
test:
	bats test/mohawk.bats

.PHONY: secret
secret:
	openssl ecparam -genkey -name secp384r1 -out server.key
	openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650 -subj /C=US/ST=name/O=comp

.PHONY: container
container:
	# systemctl start docker
	docker build -t yaacov/mohawk ./
	docker tag yaacov/mohawk docker.io/yaacov/mohawk
	# docker push docker.io/yaacov/mohawk
	# docker run --name mohawk -e HAWKULAE_BACKEND="memory" -v $(readlink -f ./):/root/ssh:Z yaacov/mohawk

.PHONY: install
install: fmt mohawk
	install -D -m0755 mohawk $(DESTDIR)$(BINDIR)/mohawk
