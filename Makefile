include Makefile.defs
ifndef VERSION
VERSION=latest
endif

all:
	docker build -t us.gcr.io/planetscale-operator/cilium-etcd-operator:${VERSION} .

cilium-etcd-operator:
	CGO_ENABLED=0 GOOS=linux go build $(GOBUILD) -a -installsuffix cgo -o $@
