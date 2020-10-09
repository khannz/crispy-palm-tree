if [ "$1" -ge 1 ]; then
  systemctl stop lbost1ah.service
fi
if [ "$1" = 0 ]; then
  systemctl stop lbost1ah.service
  systemctl disable lbost1ah.service
fi
exit 0