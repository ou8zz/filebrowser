#!/bin/sh
echo 'version:'$1
rm filebrowser
make build

echo "version:$1"
docker build -f Dockerfile.s6 -t ou88zz/filebrowser:$1 .

docker push ou88zz/filebrowser:$1