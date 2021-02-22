.DEFAULT_GOAL := run

version = `git fetch --tags >/dev/null && git describe --tags | cut -c 2-`
branch = `git branch --show-current`
test_version = "$(version)-$(branch)"
test = $(ENV)
prod_docker_container = akarasz/yahtzee
test_docker_container = $(CONTAINER)

.PHONY := build
build:
	go build ./...

.PHONY := test
test:
	go test ./...

.PHONY := docker
docker:
ifeq ($(strip $(test)),)
	docker build -t "$(prod_docker_container):latest" -t "$(prod_docker_container):$(version)" .
else
	docker build -t "$(test_docker_container):latest" -t "$(test_docker_container):$(test_version)" .
endif

.PHONY := run
run:
	REDIS=localhost:6379 RABBIT=amqp://guest:guest@localhost:5672 go run cmd/server/main.go

push: docker
ifeq ($(strip $(test)),)
	docker push $(prod_docker_container):latest
	docker push $(prod_docker_container):$(test_version)
else
	docker push $(test_docker_container):latest
	docker push $(test_docker_container):$(test_version)
endif
