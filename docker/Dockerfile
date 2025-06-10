# syntax=docker/dockerfile:1.4
########################################################################################################################
# Omnibus Image
########################################################################################################################

######## Build the frontend
FROM --platform=${BUILDPLATFORM} node AS frontendbuild
WORKDIR /go/src/github.com/hass-security/hass-security
COPY --link . /go/src/github.com/hass-security/hass-security

RUN make binary-frontend


######## Build the backend
FROM golang:1.20-bookworm as backendbuild

WORKDIR /go/src/github.com/hass-security/hass-security
COPY --link . /go/src/github.com/hass-security/hass-security
RUN apt-get update && DEBIAN_FRONTEND=noninteractive \
    apt-get install -y --no-install-recommends \
    file \
    && rm -rf /var/lib/apt/lists/*
RUN make binary-clean binary-all WEB_BINARY_NAME=hass-security


######## Combine build artifacts in runtime image
FROM debian:bookworm-slim as runtime
ARG TARGETARCH
EXPOSE 8080
WORKDIR /opt/hass-security
ENV PATH="/opt/hass-security/bin:${PATH}"
ENV INFLUXD_CONFIG_PATH=/opt/hass-security/influxdb
ENV S6VER="3.1.6.2"
ENV INFLUXVER="2.2.0"
ENV S6_SERVICES_READYTIME=1000
SHELL ["/usr/bin/sh", "-c"]

RUN apt-get update && DEBIAN_FRONTEND=noninteractive \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    cron \
    curl \
    smartmontools \
    tzdata \
    procps \
    xz-utils \
    && rm -rf /var/lib/apt/lists/* \
    && update-ca-certificates \
    && case ${TARGETARCH} in \
           "amd64")  S6_ARCH=x86_64  ;; \
           "arm64")  S6_ARCH=aarch64  ;; \
       esac \
    && curl https://github.com/just-containers/s6-overlay/releases/download/v${S6VER}/s6-overlay-noarch.tar.xz -L -s --output /tmp/s6-overlay-noarch.tar.xz \
    && tar -Jxpf /tmp/s6-overlay-noarch.tar.xz -C / \
    && rm -rf /tmp/s6-overlay-noarch.tar.xz \
    && curl https://github.com/just-containers/s6-overlay/releases/download/v${S6VER}/s6-overlay-${S6_ARCH}.tar.xz -L -s --output /tmp/s6-overlay-${S6_ARCH}.tar.xz \
    && tar -Jxpf /tmp/s6-overlay-${S6_ARCH}.tar.xz -C / \
    && rm -rf /tmp/s6-overlay-${S6_ARCH}.tar.xz
RUN curl -L https://dl.influxdata.com/influxdb/releases/influxdb2-${INFLUXVER}-${TARGETARCH}.deb --output /tmp/influxdb2-${INFLUXVER}-${TARGETARCH}.deb \
    && dpkg -i --force-all /tmp/influxdb2-${INFLUXVER}-${TARGETARCH}.deb \
    && rm -rf /tmp/influxdb2-${INFLUXVER}-${TARGETARCH}.deb

COPY /rootfs /

COPY --link --from=backendbuild --chmod=755 /go/src/github.com/hass-security/hass-security/hass-security /opt/hass-security/bin/
COPY --link --from=backendbuild --chmod=755 /go/src/github.com/hass-security/hass-security/hass-security-collector-metrics /opt/hass-security/bin/
COPY --link --from=frontendbuild --chmod=644 /go/src/github.com/hass-security/hass-security/dist /opt/hass-security/web
RUN chmod 0644 /etc/cron.d/hass-security && \
    rm -f /etc/cron.daily/* && \
    mkdir -p /opt/hass-security/web && \
    mkdir -p /opt/hass-security/config && \
    chmod -R ugo+rwx /opt/hass-security/config && \
    chmod +x /etc/cont-init.d/* && \
    chmod +x /etc/services.d/*/run && \
    chmod +x /etc/services.d/*/finish

CMD ["/init"]
