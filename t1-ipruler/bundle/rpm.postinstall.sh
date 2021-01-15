lbost1aipr="lbost1aipr"

mkdir -p "/var/run/$lbost1aipr"

systemctl daemon-reload
systemctl enable --now lbost1aipr.service
