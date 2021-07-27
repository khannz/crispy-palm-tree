if [ "$1" -ge 1 ]; then
  systemctl stop lbost1ai.service
fi
if [ "$1" = 0 ]; then
  systemctl disable --now lbost1ai.service
fi
