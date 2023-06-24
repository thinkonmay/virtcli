systemctl stop virtdaemon.service
echo "[Unit]
Description=
After=network.target

[Service]
Type=simple
ExecStart=/snap/bin/go run cmd/server/main.go
WorkingDirectory=/home/virtdaemon

[Install]
WantedBy=multi-user.target" > /lib/systemd/system/virtdaemon.service

systemctl enable virtdaemon.service
systemctl start  virtdaemon.service
systemctl status virtdaemon.service