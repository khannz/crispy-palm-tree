LBOST1AC="lbost1ac"

mkdir -p "/var/run/$LBOST1AC"

systemctl daemon-reload
systemctl enable lbost1ac.service
# systemctl start lbost1ac.service

exit 0