     GO = go
PACKAGE = github.com/atinm/spotify-filter
   GOOS = $(shell go env GOOS)
 GOARCH = $(shell go env GOARCH)

PACKAGE_LIST = \
	$(PACKAGE)/lib \
	$(PACKAGE)/desktop

all ::
	$(GO) install -v $(PACKAGE_LIST)

clean ::
	$(GO) clean -i -x $(PACKAGE_LIST)
	rm -rf $(GOPATH)/pkg/$(GOOS)_$(GOARCH)/$(PACKAGE)

wc ::
	wc -l *.go */*.go examples/*/*.go

longlines ::
	egrep '.{120,}' *.go */*.go examples/*/*.go

fmt ::
	$(GO) fmt -x $(PACKAGE_LIST)

vet ::
	$(GO) vet -x $(PACKAGE_LIST)
