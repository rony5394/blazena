FROM alpine:3

RUN mkdir -p /root/.ssh/
RUN apk add openssh rsync --no-cache
