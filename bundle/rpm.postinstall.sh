SNLB_DIR="snlb"

mkdir -p "/var/run/$SNLB_DIR"

systemctl daemon-reload
systemctl enable snlb.service
# systemctl start snlb.service

exit 0