FROM golang:1.14-alpine

RUN echo $GOPATH
RUN apk add --no-cache make gcc musl-dev git
WORKDIR $GOPATH/src/github.com/qtumproject/janus
COPY ./ $GOPATH/src/github.com/qtumproject/janus
RUN go install github.com/qtumproject/janus/cli/janus

ENV QTUM_RPC=http://qtum:testpasswd@localhost:3889
ENV QTUM_NETWORK=regtest

EXPOSE 23889

ENTRYPOINT [ "janus" ]