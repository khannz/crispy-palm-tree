lbost1at="lbost1at"

mkdir -p "/var/run/$lbost1at"

systemctl daemon-reload
systemctl enable --now lbost1at.service
