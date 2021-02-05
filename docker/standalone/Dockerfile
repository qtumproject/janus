FROM golang:1.14-alpine

RUN apk add --no-cache make musl-dev git
#I think the reason it builds the image without the new methods is because of this
RUN go get github.com/qtumproject/janus/...

ENV QTUM_RPC=http://qtum:testpasswd@localhost:3889
ENV QTUM_NETWORK=regtest

EXPOSE 23889

ENTRYPOINT [ "janus"]