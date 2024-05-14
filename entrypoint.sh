#!/bin/sh

set -e

useradd -ms /bin/bash sogo

# Set process UID and GID at runtime
if [ -n "$PUID" ] && [ -n "$PGID" ]; then
  groupmod -g $PGID sogo
  usermod -u $PUID -g $PGID sogo
fi

# create mandatory dirs and enforce owner+mode
for dir in lib log run spool; do
  install -m 750 -o sogo -g sogo -d /var/$dir/sogo
done

# Make example scripts in /usr/share/doc/sogo/ executable
# (they do not really belong there, we are violating Debian
# packaging guidelines, but OTOH moving these files now would
# break lots of setups)
if [ -d "/usr/share/doc/sogo" ] && [ $(ls -al /usr/share/doc/sogo/ | grep .sh |  wc -l) -gt 0 ]; then
  chmod a+x /usr/share/doc/sogo/*.sh
fi

# Create custom yaml config folder
mkdir -p "/etc/sogo/sogo.conf.d"

if [ -z "$(ls -A /etc/sogo/sogo.conf.d)" ]; then
  echo "/etc/sogo/sogo.conf.d is empty. Falling back to using existing /etc/sogo/sogo.conf..."
else
  # Generate config file from yaml folder
  echo "Generating sogo.conf from /etc/sogo/sogo.conf.d YAML files..."
  . /opt/config_parser.sh
  GenerateConfigFile
fi

# Enforce owner+mode on configuration
chmod 750 /etc/sogo
chown root:sogo /etc/sogo
chmod 640 /etc/sogo/sogo.conf
chown root:sogo /etc/sogo/sogo.conf
chmod +x /usr/sbin/sogod

PIDFILE=/var/run/sogo/sogo.pid
LOGFILE=/dev/fd/1

export DAEMON_OPTS="-WOPidFile $PIDFILE -WOLogFile $LOGFILE"

# Start supervisor
/usr/bin/supervisord -c /opt/supervisord.conf