go build -o virtdaemon cmd/server/main.go
sudo systemctl stop virtdaemon.service
sudo echo "[Unit]
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

sudo systemctl enable virtdaemon.service
sudo systemctl start  virtdaemon.service
sudo systemctl status virtdaemon.service