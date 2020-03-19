test:
	go test -v ./...

build: ssg
ssg:
	go build ./cmd/ssg

docker:
	@echo "Checking that VERSION was defined in the calling environment"
	@test -n "$(VERSION)"
	@echo "OK.  VERSION=$(VERSION)"
	
	docker build -t starkandwayne-private/shield-ssg:$(VERSION) . --build-arg VERSION=$(VERSION)
	docker tag starkandwayne-private/shield-ssg:$(VERSION) starkandwayne-private/shield-ssg:latest

.PHONY: test build ssg docker
