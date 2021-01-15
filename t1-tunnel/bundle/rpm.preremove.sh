if [ "$1" -ge 1 ]; then
  systemctl stop lbost1at.service
fi
if [ "$1" = 0 ]; then
  systemctl stop lbost1at.service
  systemctl disable lbost1at.service
fi
exit 0