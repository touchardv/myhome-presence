ARG ALPINE_VERSION=latest

FROM alpine:${ALPINE_VERSION}

ARG TARGETARCH

VOLUME [ "/etc/myhome", "/var/lib/myhome" ]

COPY deployment/config.yaml /etc/myhome/config.yaml
COPY build/myhome-presence-linux-${TARGETARCH} /myhome-presence

RUN addgroup -S myhome && \
	adduser -S -D -G myhome -H -s /bin/sh myhome

ENV HOME=/

USER myhome

EXPOSE 8080

ENTRYPOINT [ "/myhome-presence" ]
