#
# Build go project
#
FROM golang:1.15-alpine as go-builder

WORKDIR /go/src/github.com/in4it/ecs-ssh/

COPY . .

RUN apk add -u -t build-tools curl git make && \
    make && \
    apk del build-tools && \
    rm -rf /var/cache/apk/*

#
# Runtime container
#
FROM alpine:3.15.4  

RUN apk --no-cache add ca-certificates && \
    addgroup -g 1000 app && \
    adduser -h /app -s /bin/sh -G app -S -u 1000 app 

WORKDIR /app

COPY --from=go-builder /go/src/github.com/in4it/ecs-ssh/ecs-ssh-linux-amd64 ecs-ssh

USER app

CMD ["./ecs-ssh"]  
