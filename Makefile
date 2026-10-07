.PHONY: test vet build docker

test:
	go test ./...

vet:
	go vet ./...

build:
	go build ./cmd/week-in-review

docker:
	docker build -t week-in-review:dev .
