#!/bin/bash

if [ -f /etc/default/sogo ]; then
    . /etc/default/sogo
fi

. /lib/lsb/init-functions
. /usr/share/GNUstep/Makefiles/GNUstep.sh

/usr/local/sbin/sogod -WONoDetach YES -WOLogFile /var/log/sogo/sogo.log