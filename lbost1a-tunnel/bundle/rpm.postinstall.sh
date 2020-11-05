lbost1at="lbost1at"

mkdir -p "/var/run/$lbost1at"

systemctl daemon-reload
systemctl enable lbost1at.service
# systemctl start lbost1at.service

exit 0