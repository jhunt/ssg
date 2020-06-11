FROM ubuntu:20.04 AS certs
RUN apt-get update \
 && apt-get install -y ca-certificates
# /etc/ssl/certs/ca-certificates.crt

FROM golang:1.13 AS build

ARG VERSION
WORKDIR /app
COPY . /app

ENV CGO_ENABLED=0
RUN go build -o ssg -ldflags "-X main.Version=$VERSION" .

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /app/ssg /ssg
ENTRYPOINT ["/ssg"]
