lbost1ad="lbost1ad"
if [ "$1" -ge 1 ]; then
  echo "$1"
fi
if [ "$1" = 0 ]; then
  systemctl daemon-reload
  rm -rf "/opt/$lbost1ad"
  rm -rf "/var/run/$lbost1ad"
fi
exit 0