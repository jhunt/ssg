FROM golang:1.13 AS build

ARG VERSION
WORKDIR /app
COPY . /app

RUN go build -o ssg -ldflags "-linkmode external -extldflags -static -X main.Version=$VERSION" .

FROM scratch
COPY --from=build /app/ssg /ssg
ENTRYPOINT ["/ssg"]
