export GOPATH:=$(CURDIR)/Godeps/_workspace:$(GOPATH)

LIBDIR=${DESTDIR}/lib/systemd/system
BINDIR=${DESTDIR}/usr/lib/docker/

#PACKAGES = $(shell find ./ -type d -not -path '*/\.*')
PACKAGES=./utils

.PHONY: test-cover-html

all:
	go build  -o docker-auth-plugin .

test:
	go test ${PACKAGES}

coverage:
	echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out  $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)
	go tool cover -html=coverage-all.out -o coverage.html
	rm -rf coverage-all.out coverage.out

install:
	install -d -m 0755 ${LIBDIR}
	install -m 644 systemd/docker-auth-plugin.service ${LIBDIR}
	install -d -m 0755 ${BINDIR}
	install -m 755 docker-auth-plugin ${BINDIR}

clean:
	rm docker-auth-plugin *~
