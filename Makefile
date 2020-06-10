IMAGE ?= huntprod/ssg
TAG   ?= latest

test:
	go test -v ./...
	prove -v

build: ssg
ssg:
	go build -o ssg .

docker:
	docker build -t $(IMAGE):latest .

release:
	@echo "Checking that VERSION was defined in the calling environment"
	@test -n "$(VERSION)"
	@echo "OK.  VERSION=$(VERSION)"
	
	docker build -t $(IMAGE):$(VERSION) . --build-arg VERSION=$(VERSION)
	docker tag $(IMAGE):$(VERSION) $(IMAGE):latest
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest

coverage:
	go test -coverprofile=cover.out ./...; go tool cover -html=cover.out; rm cover.out

.PHONY: test build ssg docker
