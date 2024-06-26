#!/bin/sh
echo 'version:'$1
rm filebrowser
make build

echo "version:$1"
docker build -f Dockerfile -t registry.cn-shanghai.aliyuncs.com/ou88zz/filebrowser:$1 .

# docker push ou88zz/filebrowser:$1
docker push registry.cn-shanghai.aliyuncs.com/ou88zz/filebrowser:$1