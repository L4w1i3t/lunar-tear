FROM alpine:latest AS builder

WORKDIR /usr/local/src
COPY . .

RUN apk add --no-cache \
    protobuf \
    protobuf-dev \
    protoc \
    make \
    go \
    libcap

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest &&\
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest &&\ 
    PATH="$PATH:$(go env GOPATH)/bin" make proto &&\
    go build -o lunar-tear ./cmd/lunar-tear &&\
    setcap cap_net_bind_service=+ep ./lunar-tear

FROM alpine:latest

WORKDIR /opt/lunar-tear

RUN chown 1000:1000 /opt/lunar-tear

USER 1000

COPY --from=builder /usr/local/src/lunar-tear .

COPY entrypoint.sh .

ENTRYPOINT /opt/lunar-tear/entrypoint.sh

