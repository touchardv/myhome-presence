ARG ALPINE_VERSION

FROM alpine:${ALPINE_VERSION}

ARG TARGETARCH

VOLUME [ "/etc/myhome", "/var/lib/myhome" ]

COPY deployment/config.yaml /etc/myhome/config.yaml
COPY build/myhome-presence-linux-${TARGETARCH} /myhome-presence

EXPOSE 8080

ENTRYPOINT [ "/myhome-presence" ]
