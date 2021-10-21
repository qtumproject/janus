#!/bin/sh

if [ -f "./https/cert.pem" ]; then
    echo "cert.pem already exists"
    if [ ! -f "./https/key.pem" ]; then
        echo "key.pem does not exist, both are needed, please delete "`pwd`"/https/key.pem manually"
        exit 1
    fi
    exit 0
fi

mkdir -p ./https
if [ ! -d "./https" ]; then
    echo "Failed to mkdir -p ./https"
    exit 1
fi

echo "Generating key.pem and cert.pem"
docker run -v `pwd`/https:/https qtum/openssl.janus openssl req -nodes  -x509 -newkey rsa:4096 -keyout /https/key.pem -out /https/cert.pem -days 365 -subj "/C=US/ST=ST/L=L/O=Janus Self-signed https/OU=Janus Self-signed https/CN=Janus Self-signed https"
if [ 0 -ne $? ]; then
    echo "Failed to generate server.key"
    exit $?
fi

if [ ! -r ./https/key.key ]; then
    echo "Generated files have wrong owner, chowning them"
    sudo chown `id -u`:`id -g` ./https/key.pem ./https/cert.pem
    if [ ! -r ./https/key.pem ]; then
        echo "Failed to chown on generated files, do it manually in "`pwd`"/https"
        echo "\"sudo chown `id -u`:`id -g` ./https/key.pem ./https/cert.pem\""
        exit 1
    else
        echo "Successfully chown'd generated files are now readable"
    fi
fi

exit 0

echo "Generating server.key"
docker run -v `pwd`/https:/https qtum/openssl.janus openssl genrsa -out /https/server.key 2048
if [ 0 -ne $? ]; then
    echo "Failed to generate server.key"
    exit $?
fi
docker run -v `pwd`/https:/https qtum/openssl.janus openssl ecparam -genkey -name secp384r1 -out /https/server.key
if [ 0 -ne $? ]; then
    echo "Failed to generate sever.key"
    exit $?
fi
echo "Generating server.crt"
docker run -v `pwd`/https:/https qtum/openssl.janus openssl req -nodes -new -x509 -sha256 -key /https/server.key -out /https/server.crt -days 3650 -subj "/C=US/ST=ST/L=L/O=Janus Self-signed https/OU=Janus Self-signed https/CN=Janus Self-signed https"
if [ 0 -ne $? ]; then
    echo "Failed to generate server.crt"
    exit $?
fi


