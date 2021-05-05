.PHONY: image
PROG_TAG ?= v0.0.1
BASE_TAG ?= v0.0.0
image: 
	docker build -t cxinsect/prom:$(PROG_TAG) -f Dockerfile .
	docker push cxinsect/prom:$(PROG_TAG)
clean: 
	docker rmi -f cxinsect/prom:$(PROG_TAG)

image_base:
	docker build -t cxinsect/gorilla:$(BASE_TAG) -f Dockerfile_base .
	docker push cxinsect/gorilla:$(BASE_TAG)

