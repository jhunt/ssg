FROM golang:1.13 AS build

WORKDIR /app
COPY . /app

RUN go build -ldflags "-linkmode external -extldflags -static" ./cmd/ssg

FROM scratch
COPY --from=build /app/ssg /ssg
ENTRYPOINT ["/ssg"]
