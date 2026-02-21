FROM docker.io/library/golang:1.25.5-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . /build


RUN CGO_ENABLED=0 GOOS=linux go build -o /blazena


FROM docker.io/library/alpine:3.3

COPY --from=builder /blazena /
EXPOSE 1234 
ENV MODE=invalid

CMD /blazena $MODE
