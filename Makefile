.DEFAULT_GOAL := run

.PHONY := build
build:
	go build ./...

.PHONY := test
test:
	go test --count=1 ./...

.PHONY := docker
docker:
	docker build -t akarasz/yahtzee:latest .

.PHONY := run
run: docker
	docker run -p 8000:8000 akarasz/yahtzee:latest

push: docker
	docker push akarasz/yahtzee:latest
