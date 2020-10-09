if [ "$1" -ge 1 ]; then
  systemctl stop lbost1ac.service
fi
if [ "$1" = 0 ]; then
  systemctl stop lbost1ac.service
  systemctl disable lbost1ac.service
fi
exit 0