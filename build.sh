#!/bin/sh
rm filebrowser
make build

docker build -f Dockerfile.s6 -t ou88zz/filebrowser:$0 .

docker push ou88zz/filebrowser:$0