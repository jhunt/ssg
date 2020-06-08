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

coverage:
	go test -coverprofile=cover.out ./...; go tool cover -html=cover.out; rm cover.out

.PHONY: test build ssg docker
