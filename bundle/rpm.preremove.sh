if [ "$1" -ge 1 ]; then
  systemctl stop snlb.service
fi
if [ "$1" = 0 ]; then
  systemctl stop snlb.service
  systemctl disable snlb.service
fi
exit 0