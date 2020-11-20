lbost1ad="lbost1ad"

mkdir -p "/var/run/$lbost1ad"

systemctl daemon-reload
systemctl enable --now lbost1ad.service
