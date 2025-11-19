FROM golang:1.24

WORKDIR ${GOPATH}/pullrequest-manager/
COPY . ${GOPATH}/pullrequest-manager

RUN go build -o /build ./cmd/pullrequest-manager \
        && go clean -cache -modcache

EXPOSE 8080

CMD ["/build"]