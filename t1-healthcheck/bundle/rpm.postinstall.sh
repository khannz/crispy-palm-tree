lbost1ah="lbost1ah"

mkdir -p "/var/run/$lbost1ah"

systemctl daemon-reload
systemctl enable lbost1ah.service
# systemctl start lbost1ah.service

exit 0