FROM golang:1.21
WORKDIR /root
COPY . .
RUN CGO_ENABLED=0 go build -o /tmp/s3gw

# Docker is used as a base image so you can easily start playing around in the container using the Docker command line client.
FROM docker
COPY --from=0 /tmp/s3gw /usr/local/bin/s3gw
RUN chmod +x /usr/local/bin/s3gw
RUN apk add bash curl
