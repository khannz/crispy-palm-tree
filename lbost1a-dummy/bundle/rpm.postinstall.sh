lbost1ad="lbost1ad"

mkdir -p "/var/run/$lbost1ad"

systemctl daemon-reload
systemctl enable lbost1ad.service
# systemctl start lbost1ad.service

exit 0