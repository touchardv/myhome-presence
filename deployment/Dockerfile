FROM alpine:3.20

ARG TARGETARCH
ARG BUILD_DIR=build
ARG BINARY=myhome-presence-linux-${TARGETARCH}

VOLUME [ "/etc/myhome", "/var/lib/myhome" ]

COPY deployment/config.yaml /etc/myhome/config.yaml
COPY ${BUILD_DIR}/${BINARY} /myhome-presence

RUN addgroup -S myhome && \
	adduser -S -D -G myhome -H -s /bin/sh myhome

ENV HOME=/

USER myhome

EXPOSE 8080

ENTRYPOINT [ "/myhome-presence" ]
