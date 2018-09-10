BINARY = ecs-ssh
GOARCH = amd64


all: build

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-linux-${GOARCH} . ; 

clean:
	rm -f ${BINARY}-linux-${GOARCH}
