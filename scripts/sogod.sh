#!/bin/bash

if [ -z "$LD_PRELOAD" ]; then
  LIBSSL_LOCATION=$(find / -type f -name "libssl.so.*" -print -quit 2>/dev/null)
  echo "LD_PRELOAD=$LIBSSL_LOCATION" >> /etc/default/sogo
  echo "LD_LIBRARY_PATH=/usr/local/lib/sogo:$LD_LIBRARY_PATH" >> /etc/default/sogo
  export LD_PRELOAD=$LIBSSL_LOCATION
else
  echo "LD_PRELOAD=$LD_PRELOAD" >> /etc/default/sogo
  if [ -z "$LD_LIBRARY_PATH" ]; then
    echo "LD_LIBRARY_PATH=/usr/local/lib/sogo:/usr/local/lib:/usr/lib" >> /etc/default/sogo
  else
    echo "LD_LIBRARY_PATH=/usr/local/lib/sogo:/usr/local/lib:$LD_LIBRARY_PATH" >> /etc/default/sogo
  fi
  export LD_PRELOAD=$LD_PRELOAD
fi

if [ -f /etc/default/sogo ]; then
    . /etc/default/sogo
fi

. /usr/share/GNUstep/Makefiles/GNUstep.sh

# Run SOGo in foreground
su -s /bin/sh -c '/usr/local/sbin/sogod -WONoDetach YES -WOLogFile /var/log/sogo/sogo.log' sogo