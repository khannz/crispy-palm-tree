lbost1ar="lbost1ar"

mkdir -p "/var/run/$lbost1ar"

systemctl daemon-reload
systemctl enable --now lbost1ar.service
