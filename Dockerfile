#
# Build go project
#
FROM golang:1.11-alpine as go-builder

WORKDIR /go/src/github.com/in4it/ecs-ssh/

COPY . .

RUN apk add -u -t build-tools curl git make && \
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh && \
    dep ensure && \
    make && \
    apk del build-tools && \
    rm -rf /var/cache/apk/*

#
# Runtime container
#
FROM alpine:latest  

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=go-builder /go/src/github.com/in4it/ecs-ssh/ecs-ssh-linux-amd64 ecs-ssh

CMD ["./ecs-ssh"]  
