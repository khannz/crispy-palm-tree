lbost1at="lbost1at"
if [ "$1" -ge 1 ]; then
  echo "$1"
fi
if [ "$1" = 0 ]; then
  systemctl daemon-reload
  rm -rf "/opt/$lbost1at"
  rm -rf "/var/run/$lbost1at"
fi
exit 0