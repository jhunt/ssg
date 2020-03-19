FROM golang:1.13 AS build

ARG VERSION
WORKDIR /app
COPY . /app

RUN go build -ldflags "-linkmode external -extldflags -static -X main.Version=$VERSION" ./cmd/ssg

FROM scratch
COPY --from=build /app/ssg /ssg
ENTRYPOINT ["/ssg"]
