if [ "$1" -ge 1 ]; then
  systemctl stop lbost1ad.service
fi
if [ "$1" = 0 ]; then
  systemctl stop lbost1ad.service
  systemctl disable lbost1ad.service
fi
exit 0