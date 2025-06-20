FROM alpine:3.22

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# RUN apk update && \
#   apk --no-cache add ca-certificates mailcap curl jq tini

# # Make user and create necessary directories
ENV UID=99
ENV GID=100

RUN addgroup -g $GID user && \
  adduser -D -u $UID -G user user && \
  mkdir -p /config /database /srv && \
  chown -R user:user /config /database /srv

# Copy files and set permissions
COPY filebrowser /bin/filebrowser
RUN chmod +x /bin/filebrowser
COPY docker/common/ /
COPY docker/alpine/ /

RUN chown -R user:user /bin/filebrowser /defaults healthcheck.sh init.sh

# Define healthcheck script
HEALTHCHECK --start-period=2s --interval=5s --timeout=3s CMD /healthcheck.sh

# Set the user, volumes and exposed ports
#USER user

VOLUME /srv /config /database

EXPOSE 80

CMD [ "filebrowser" ]
