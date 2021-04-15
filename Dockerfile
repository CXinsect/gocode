FROM cxinsect/gorilla:v0.0.0
	
ENV PATH="$PATH:/usr/local/go/bin"

COPY ./ /home/go/exporter

EXPOSE 8888

RUN cd /home/go/exporter/ && GOPROXY="https://goproxy.cn,direct" go build -o prom_exporter cmd/main.go

WORKDIR /home/go/exporter/


#ENTRYPOINT ["/bin/bash","-c","./prom_exporter --config.file example/brkfault.yaml"]







