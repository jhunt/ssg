test:
	go test -v ./...

build: ssg
ssg:
	go build ./cmd/ssg

docker:
	docker build -t starkandwayne-private/shield-ssg:latest .

.PHONY: test build ssg docker
