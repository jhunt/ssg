FROM golang:1.13 AS build

ARG VERSION
WORKDIR /app
COPY . /app

ENV CGO_ENABLED=0
RUN go build -o ssg -ldflags "-X main.Version=$VERSION" .

FROM scratch
COPY --from=build /app/ssg /ssg
ENTRYPOINT ["/ssg"]
