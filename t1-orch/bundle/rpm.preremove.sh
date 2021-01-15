if [ "$1" -ge 1 ]; then
  systemctl stop lbost1ao.service
fi
if [ "$1" = 0 ]; then
  systemctl stop lbost1ao.service
  systemctl disable lbost1ao.service
fi
exit 0