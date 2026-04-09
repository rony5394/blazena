FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.25.5-alpine AS builder

ARG TARGETARCH
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . /build


RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -o /blazena


FROM docker.io/library/alpine:3.23
RUN apk add openssh rsync btrfs-progs --no-cache

COPY --from=builder --chmod=+x /blazena /
EXPOSE 1234 
ENV MODE=invalid

WORKDIR /root/.ssh

CMD ["/blazena", "$MODE"]
