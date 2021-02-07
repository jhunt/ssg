FROM ubuntu:20.04 AS certs
RUN apt-get update \
 && apt-get install -y ca-certificates
# /etc/ssl/certs/ca-certificates.crt

FROM golang:1.15 AS build

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
ENV CGO_ENABLED=0
ARG VERSION
RUN go build -o ssg -ldflags "-X main.Version=$VERSION" .

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /app/ssg /ssg
ENTRYPOINT ["/ssg"]
