lbost1ao="lbost1ao"
if [ "$1" -ge 1 ]; then
  echo "$1"
fi
if [ "$1" = 0 ]; then
  systemctl daemon-reload
  rm -rf "/opt/$lbost1ao"
  rm -rf "/var/run/$lbost1ao"
fi
exit 0