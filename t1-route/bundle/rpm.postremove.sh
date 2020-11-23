lbost1ar="lbost1ar"
if [ "$1" -ge 1 ]; then
  echo "$1"
fi
if [ "$1" = 0 ]; then
  systemctl daemon-reload
  rm -rf "/opt/$lbost1ar"
  rm -rf "/var/run/$lbost1ar"
fi
exit 0