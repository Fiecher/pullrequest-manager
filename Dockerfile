FROM golang:1.24

WORKDIR ${GOPATH}/pullrequest-inator/
COPY . ${GOPATH}/pullrequest-inator

RUN go build -o /build ./cmd/pullrequest-inator \
        && go clean -cache -modcache

EXPOSE 8080

CMD ["/build"]