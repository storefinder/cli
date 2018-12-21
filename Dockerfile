FROM golang:latest as builder

ARG OS=darwin 
ARG ARCH=amd64 

WORKDIR /go/src/github.com/storefinder/cli

COPY . .

RUN go get -u golang.org/x/vgo

RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }

RUN vgo test $(vgo list ./...) -cover \
    && VERSION=$(git describe --all --exact-match `git rev-parse HEAD` | grep tags | sed 's/tags\///') \
    && GIT_COMMIT=$(git rev-list -1 HEAD) \
    && CGO_ENABLED=0 GOOS=${OS} GOARCH=${ARCH} vgo build --ldflags "-s -w \
    -X github.com/storefinder/cli/version.GitCommit=${GIT_COMMIT} \
    -X github.com/storefinder/cli/version.Version=${VERSION}" \
    -a -installsuffix cgo -o storefinder 


FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /go/src/github.com/storefinder/cli/storefinder /usr/bin/ 

ENV PATH=$PATH:/usr/bin/

CMD [ "storefinder -v" ]