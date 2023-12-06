FROM ghcr.io/linuxserver/baseimage-debian:bookworm

RUN mkdir /etc/sogo && chown -R abc: /etc/sogo

# install operating system packages
RUN apt-get update -y && apt-get install wget make git gettext gnupg2 -y
RUN wget -O- "https://keyserver.ubuntu.com/pks/lookup?op=get&search=0xCB2D3A2AA0030E2C" | gpg --dearmor | apt-key add -
RUN wget -O- "https://keys.openpgp.org/vks/v1/by-fingerprint/74FFC6D72B925A34B5D356BDF8A27B36A6E2EAE9" | gpg --dearmor | apt-key add -
RUN apt-get update && \
    apt-get install apt-transport-https -y && \
    echo "deb http://packages.sogo.nu/nightly/5/debian/ bookworm bookworm" >> /etc/apt/sources.list && \
    apt-get update && \
    apt-get install sogo sope4.9-gdl1-postgresql sope4.9-gdl1-mysql default-mysql-client nginx -y && \
    rm -rf /var/lib/apt/lists/*

COPY ./bashutil/* /usr/local/bin

# add config and init files
ADD config /opt/docker-config
ADD init /opt/docker-init

RUN chmod +x /opt/docker-init/entrypoint && \
    chmod +x /usr/local/bin/bgo && \
    chmod +x /usr/local/bin/bgowait && \
    chmod +x /usr/local/bin/retry

# start from init folder
WORKDIR /opt/docker-init
ENTRYPOINT ["/opt/docker-init/entrypoint"]
