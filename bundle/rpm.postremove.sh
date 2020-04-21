SNLB_DIR="snlb"
if [ "$1" -ge 1 ]; then
  echo "$1"
fi
if [ "$1" = 0 ]; then
  systemctl daemon-reload
  rm -rf "/opt/$SNLB_DIR"
  rm -rf "/var/run/$SNLB_DIR"
fi
exit 0