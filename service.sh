go build -o virtdaemon cmd/server/main.go
systemctl stop virtdaemon.service
echo "[Unit]
Description=
After=network.target

StartLimitIntervalSec=500
StartLimitBurst=5


[Service]
Type=simple
ExecStart=/home/virtdaemon/virtdaemon
WorkingDirectory=/home/virtdaemon

Restart=always
RestartSec=5s


[Install]
WantedBy=multi-user.target" > /lib/systemd/system/virtdaemon.service

systemctl enable virtdaemon.service
systemctl start  virtdaemon.service
systemctl status virtdaemon.service