LBOST1AC="lbost1ac"
if [ "$1" -ge 1 ]; then
  echo "$1"
fi
if [ "$1" = 0 ]; then
  systemctl daemon-reload
  rm -rf "/opt/$LBOST1AC"
  rm -rf "/var/run/$LBOST1AC"
fi
exit 0