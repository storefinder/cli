FROM golang:latest as builder

WORKDIR /github.com/storefinder/cli

COPY . .

RUN go get -u golang.org/x/vgo

RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }

RUN vgo test $(vgo list ./.. ) \
    && CGO_ENABLED=0 GOOS=linux vgo build -a -o storefinder . 
 
FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /github.com/storefinder/cli/storefinder .

CMD [ "./"storefinder" ]