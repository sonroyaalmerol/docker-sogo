FROM debian:bookworm AS builder

ARG SOGO_VERSION

ADD https://packages.sogo.nu/sources/SOGo-${SOGO_VERSION}.tar.gz /tmp/SOGo.tar.gz
ADD https://packages.sogo.nu/sources/SOPE-${SOGO_VERSION}.tar.gz /tmp/SOPE.tar.gz

RUN apt-get update -y && \
    apt-get install -y \
        git \
        zip \
        wget \
        make \
        gnustep-make \
        gnustep-base-common \
        gnustep-base-runtime \
        libgnustep-base-dev \
        gobjc \
        libwbxml2-dev \
        libxml2-dev \
        libssl-dev \
        libldap-dev \
        libpq-dev \
        libmemcached-dev \
        default-libmysqlclient-dev \
        libytnef0-dev \
        zlib1g-dev \
        liblasso3-dev \
        libcurl4-gnutls-dev \
        libexpat1-dev \
        libpopt-dev \
        libsbjson-dev \
        libsbjson2.3 \
        libcurl4 \
        liboath-dev \
        libsodium-dev \
        libzip-dev && \
    rm -rf /var/lib/apt/lists/* && \
    mkdir /tmp/SOGo && \
    mkdir /tmp/SOPE && \
    tar xf /tmp/SOGo.tar.gz -C /tmp/SOGo --strip-components 1 && \
    tar xf /tmp/SOPE.tar.gz -C /tmp/SOPE --strip-components 1 && \
    cd /tmp/SOPE && \
    ./configure --with-gnustep --enable-debug --disable-strip && \
    make && \
    make install && \
    cd /tmp/SOGo && \
    ./configure --enable-debug --disable-strip && \
    make && \
    make install

FROM debian:bookworm-slim

ARG TARGETARCH

ENV PUID=1000
ENV PGID=1000

# install dependencies

RUN apt-get update -y && \
    apt-get install -y \
        ca-certificates \
        tzdata \
        wget \
        make \
        git \
        cron \
        gettext \
        gnupg2 \
        default-mysql-client \
        postgresql-client \
        apache2 \
        supervisor \
        gnustep-base-runtime \
        libc6 \
        libcrypt1 \
        liblasso3 \
        lsb-base \
        libwbxml2-1 \
        libcurl4 \
        libgcc-s1 \
        libglib2.0-0 \
        libgnustep-base1.28 \
        libldap-2.5-0 \
        libmemcached11 \
        libsodium23 \
        libzip4 \
        liboath0 \
        libobjc4 \
        libpq5 \
        libssl3 \
        libxml2 \
        libsope1 \
        libsbjson2.3 \
        libytnef0 \
        zlib1g \
        postgresql-client-common \
        postgresql-common \
        mysql-common && \
    rm -rf /var/lib/apt/lists/*

# add config, binaries, libraries, and init files
COPY --from=builder /usr/local/sbin/ /usr/local/sbin/
COPY --from=builder /usr/local/lib/sogo/ /usr/local/lib/sogo/
COPY --from=builder /usr/local/lib/GNUstep/ /usr/local/lib/GNUstep/
COPY --from=builder /usr/local/include/GNUstep/ /usr/local/include/GNUstep/
COPY --from=builder /usr/share/GNUstep/Makefiles/ /usr/share/GNUstep/Makefiles/
COPY --from=builder /etc/GNUstep/ /etc/GNUstep/
COPY --from=builder /tmp/SOGo/Scripts/sogo-default /etc/default/sogo
COPY --from=builder /tmp/SOGo/Scripts/sogo.cron /etc/cron.d/sogo
COPY --from=builder /tmp/SOGo/Scripts/sogo.conf /etc/sogo/sogo.conf
COPY --from=builder /tmp/SOGo/Scripts/ /usr/share/doc/sogo/
COPY --from=builder /tmp/SOGO/Apache/SOGo.conf /etc/apache2/conf-available/SOGo.conf

COPY supervisord.conf /opt/supervisord.conf
COPY config_parser.sh /opt/config_parser.sh
COPY entrypoint.sh /opt/entrypoint.sh

ADD https://github.com/mikefarah/yq/releases/latest/download/yq_linux_${TARGETARCH} /usr/bin/yq

RUN echo "/usr/local/lib/sogo" > /etc/ld.so.conf.d/sogo.conf && \
    ldconfig && \
    groupadd --system sogo && \
    useradd --system --gid sogo sogo && \
    mkdir -p /usr/lib/GNUstep/ && \
    ln -s /usr/local/lib/GNUstep/SOGo /usr/lib/GNUstep/SOGo && \
    ln -s /usr/local/sbin/sogo-tool /usr/sbin/sogo-tool && \
    ln -s /usr/local/sbin/sogo-ealarms-notify /usr/sbin/sogo-ealarms-notify && \
    ln -s /usr/local/sbin/sogo-slapd-sockd /usr/sbin/sogo-slapd-sockd && \
    chmod +rx /usr/bin/yq && \
    chmod +rx /opt/entrypoint.sh

# start from config folder
WORKDIR /etc/sogo

ENTRYPOINT ["/opt/entrypoint.sh"]
