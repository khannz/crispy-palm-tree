lbost1ai="lbost1ai"

mkdir -p "/var/run/$lbost1ai"

systemctl daemon-reload
systemctl enable lbost1ai.service
# systemctl start lbost1ai.service

exit 0