.PHONY: image

image: 
	docker build -t cxinsect/prom:v0.0.0 -f Dockerfile .
clean:
	docker rmi -f cxinsect/prom:v0.0.0 

image_base:
	docker build -t cxinsect/gorilla:v0.0.0 -f Dockerfile_base .
	docker push cxinsect/gorilla:v0.0.0

