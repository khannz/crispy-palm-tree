lbost1ao="lbost1ao"

mkdir -p "/var/run/$lbost1ao"

systemctl daemon-reload
systemctl enable lbost1ao.service
# systemctl start lbost1ao.service

exit 0