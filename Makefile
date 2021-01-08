.DEFAULT_GOAL := run

version = `git fetch --tags >/dev/null && git describe --tags | cut -c 2-`
docker_container = akarasz/yahtzee
docker_tags = $(version),latest

.PHONY := build
build:
	go build ./...

.PHONY := test
test:
	go test ./...

.PHONY := docker
docker:
	docker build -t "$(docker_container):latest" -t "$(docker_container):$(version)" .

.PHONY := run
run: docker
	docker run -p 8000:8000 "$(docker_container):latest"

push: docker
	docker push $(docker_container):latest
	docker push $(docker_container):$(version)
