if [ "$1" -ge 1 ]; then
  systemctl stop lbost1aipr.service
fi
if [ "$1" = 0 ]; then
  systemctl stop lbost1aipr.service
  systemctl disable lbost1aipr.service
fi
exit 0