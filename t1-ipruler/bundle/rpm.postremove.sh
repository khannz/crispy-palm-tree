lbost1aipr="lbost1aipr"
if [ "$1" -ge 1 ]; then
  echo "$1"
fi
if [ "$1" = 0 ]; then
  systemctl daemon-reload
  rm -rf "/opt/$lbost1aipr"
  rm -rf "/var/run/$lbost1aipr"
fi
exit 0