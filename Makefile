VERSION				:=	$(shell cat ./VERSION)
ROOTFS_IMAGE		:=	xfsvol-rootfs
ROOTFS_CONTAINER	:=	rootfs
PLUGIN_NAME			:=	xfsvol
PLUGIN_FULL_NAME	:=	049736579808.dkr.ecr.eu-west-1.amazonaws.com/docker-plugins/xfsvol


all: install


install:
	cd ./xfsvolctl && \
		go install \
			-ldflags "-X main.version=$(VERSION)" \
			-v
	cd ./plugin && \
		go install \
			-ldflags "-X main.version=$(VERSION)" \
			-v


test:
	go test ./... -v


fmt:
	go fmt ./...
	find ./xfs -name "*.c" -o -name "*.h" | \
		xargs clang-format -style=file -i


rootfs-image:
	docker build -t $(ROOTFS_IMAGE) .


rootfs: rootfs-image
	docker rm -vf $(ROOTFS_CONTAINER) || true
	docker create --name $(ROOTFS_CONTAINER) $(ROOTFS_IMAGE) || true
	mkdir -p plugin/rootfs
	rm -rf plugin/rootfs/*
	docker export $(ROOTFS_CONTAINER) | tar -x -C plugin/rootfs
	docker rm -vf $(ROOTFS_CONTAINER)


plugin: rootfs
	docker plugin disable $(PLUGIN_NAME) || true
	docker plugin rm --force $(PLUGIN_NAME) || true
	docker plugin create $(PLUGIN_NAME) ./plugin
	docker plugin enable $(PLUGIN_NAME) || true


plugin-push: rootfs
	docker plugin rm --force $(PLUGIN_FULL_NAME) || true
	docker plugin create $(PLUGIN_FULL_NAME) ./plugin
	docker plugin create $(PLUGIN_FULL_NAME):v$(VERSION) ./plugin
	docker plugin push $(PLUGIN_FULL_NAME)
	docker plugin push $(PLUGIN_FULL_NAME):v$(VERSION)


.PHONY: install test fmt rootfs-image rootfs plugin plugin-push
