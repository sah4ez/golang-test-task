FROM golang:1.9

RUN apt-get update && apt-get install -y curl && curl https://glide.sh/get | sh
ENV GOPATH=/go

COPY glide.* /go/src/github.com/sah4ez/golang-test-task/
WORKDIR /go/src/github.com/sah4ez/golang-test-task
RUN glide install

COPY . /go/src/github.com/sah4ez/golang-test-task

RUN mkdir -p /go/bin && go build -o /go/bin/gotest && chmod +x /go/bin/gotest 

ENTRYPOINT ["/go/bin/gotest"]
EXPOSE 9990
