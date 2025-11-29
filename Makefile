BINARY=apigateway

.PHONY: build run docker-build clean

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY) -config config.yaml

docker-build:
	docker build -t cs02/apigateway:latest .

clean:
	rm -f $(BINARY)
