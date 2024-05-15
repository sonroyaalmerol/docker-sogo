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
        cmake \
        python3 \
        python-is-python3 \
        gnustep-make \
        gnustep-base-common \
        gnustep-base-runtime \
        libgnustep-base-dev \
        gobjc \
        libwbxml2-dev \
        libxml2-dev \
        libxml2-utils \
        libssl-dev \
        libldap2-dev \
        libpq-dev \
        libmemcached-dev \
        libmariadb-dev-compat \
        libytnef0-dev \
        zlib1g-dev \
        liblasso3-dev \
        libcurl4-openssl-dev \
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
    ./configure --enable-debug --disable-strip --enable-saml2 \
        --enable-mfa --enable-sodium && \
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
        zlib1g && \
    rm -rf /var/lib/apt/lists/* &&\
    curl -L -o /usr/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_$TARGETARCH

# add config, binaries, libraries, and init files
COPY --from=builder /usr/local/sbin/ /usr/local/sbin/
COPY --from=builder /usr/local/lib/ /usr/local/lib/
COPY --from=builder /tmp/SOGo/Scripts/ /usr/share/doc/sogo/
COPY --from=builder /tmp/SOGo/Apache/SOGo.conf /etc/apache2/conf-available/SOGo.conf

COPY scripts/ /opt/

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
    (ln -s /usr/local/lib/GNUstep/* /usr/lib/GNUstep/ || :) && \
    ln -s /usr/local/lib/GNUstep/Libraries/Resources /usr/lib/GNUstep/Libraries/Resources && \
    ln -s /usr/local/sbin/sogo-tool /usr/sbin/sogo-tool && \
    ln -s /usr/local/sbin/sogo-ealarms-notify /usr/sbin/sogo-ealarms-notify && \
    ln -s /usr/local/sbin/sogo-slapd-sockd /usr/sbin/sogo-slapd-sockd && \
    ln -s /etc/apache2/conf-available/SOGo.conf /etc/apache2/conf-enabled/SOGo.conf && \
    mkdir -p /etc/cron.d /etc/default /etc/sogo /etc/logrotate.d && \
    mv /usr/share/doc/sogo/sogo.cron /etc/cron.d/sogo && \
    mv /usr/share/doc/sogo/sogo-default /etc/default/sogo && \
    mv /usr/share/doc/sogo/sogo.conf /etc/sogo/sogo.conf && \
    mv /usr/share/doc/sogo/logrotate /etc/logrotate.d/sogo && \
    chmod +rx /usr/bin/yq && \
    chmod +rx /opt/entrypoint.sh && \
    chmod +rx /opt/sogod.sh

# start from config folder
WORKDIR /etc/sogo

# test test

ENTRYPOINT ["/opt/entrypoint.sh"]
