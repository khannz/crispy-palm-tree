if [ "$1" -ge 1 ]; then
  systemctl stop lbost1ar.service
fi
if [ "$1" = 0 ]; then
  systemctl stop lbost1ar.service
  systemctl disable lbost1ar.service
fi
exit 0