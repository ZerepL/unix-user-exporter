.PHONY: build run docker-build docker-run clean

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
BINARY_NAME=unix-user-exporter

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

run: build
	./$(BINARY_NAME)

docker-build:
	docker build -t $(BINARY_NAME) .

docker-run: docker-build
	docker run -p 32142:32142 -v /var/run/utmp:/var/run/utmp:ro $(BINARY_NAME)

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
