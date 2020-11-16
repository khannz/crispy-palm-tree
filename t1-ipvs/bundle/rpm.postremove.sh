lbost1ai="lbost1ai"
if [ "$1" -ge 1 ]; then
  echo "$1"
fi
if [ "$1" = 0 ]; then
  systemctl daemon-reload
  rm -rf "/opt/$lbost1ai"
  rm -rf "/var/run/$lbost1ai"
fi
exit 0