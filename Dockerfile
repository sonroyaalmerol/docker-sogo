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
        curl \
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
        libcurl4 \
        libgcc-s1 \
        libglib2.0-0 \
        libgnustep-base1.28 \
        libgnutls30 \
        liblasso3 \
        libldap-2.5-0 \
        libmariadb3 \
        libmemcached11 \
        liboath0 \
        libobjc4 \
        libpq5 \
        libsbjson2.3 \
        libsodium23 \
        libssl3 \
        libwbxml2-1 \
        libxml2 \
        libytnef0 \
        libzip4 \
        tmpreaper \
        zip \
        zlib1g \
        postgresql-client-common \
        postgresql-common \
        mysql-common && \
    rm -rf /var/lib/apt/lists/*

# add config, binaries, libraries, and init files
COPY --from=builder /usr/local/sbin/ /usr/local/sbin/
COPY --from=builder /usr/local/lib/*.so /usr/local/lib/
COPY --from=builder /usr/local/lib/GNUstep/ /usr/local/lib/GNUstep/
COPY --from=builder /usr/local/lib/sogo/*.so /usr/local/lib/sogo/
COPY --from=builder /usr/local/include/GNUstep/ /usr/local/include/GNUstep/
COPY --from=builder /usr/share/GNUstep/Makefiles/ /usr/share/GNUstep/Makefiles/
COPY --from=builder /etc/GNUstep/ /etc/GNUstep/
COPY --from=builder /tmp/SOGo/Scripts/sogo-default /etc/default/sogo
COPY --from=builder /tmp/SOGo/Scripts/sogo.cron /etc/cron.d/sogo
COPY --from=builder /tmp/SOGo/Scripts/sogo.conf /etc/sogo/sogo.conf
COPY --from=builder /tmp/SOGo/Scripts/ /usr/share/doc/sogo/
COPY --from=builder /tmp/SOGo/Apache/SOGo.conf /etc/apache2/conf-available/SOGo.conf

COPY supervisord.conf /opt/supervisord.conf
COPY config_parser.sh /opt/config_parser.sh
COPY entrypoint.sh /opt/entrypoint.sh

ADD https://github.com/mikefarah/yq/releases/latest/download/yq_linux_${TARGETARCH} /usr/bin/yq

RUN a2enmod \
        headers \
        proxy \
        proxy_http \
        rewrite \
        ssl && \
    echo "/usr/local/lib/sogo" > /etc/ld.so.conf.d/sogo.conf && \
    ldconfig && \
    groupadd --system sogo && \
    useradd --system --gid sogo sogo && \
    mkdir -p /usr/lib/GNUstep/ && \
    (ln -s /usr/local/lib/*.so /usr/lib/ || :) && \
    (ln -s /usr/local/lib/GNUstep/* /usr/lib/GNUstep/ || :) && \
    ln -s /usr/local/lib/GNUstep/Libraries/Resources /usr/lib/GNUstep/Libraries/Resources && \
    ln -s /usr/local/include/GNUstep /usr/include/GNUstep && \
    ln -s /usr/local/lib/sogo /usr/lib/sogo && \
    ln -s /usr/local/sbin/sogo-tool /usr/sbin/sogo-tool && \
    ln -s /usr/local/sbin/sogo-ealarms-notify /usr/sbin/sogo-ealarms-notify && \
    ln -s /usr/local/sbin/sogo-slapd-sockd /usr/sbin/sogo-slapd-sockd && \
    ln -s /etc/apache2/conf-available/SOGo.conf /etc/apache2/conf-enabled/SOGo.conf && \
    chmod +rx /usr/bin/yq && \
    chmod +rx /opt/entrypoint.sh

# start from config folder
WORKDIR /etc/sogo

ENTRYPOINT ["/opt/entrypoint.sh"]
