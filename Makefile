SERVICE_NAME=cuba
SERVICE_FILE=/etc/systemd/system/$(SERVICE_NAME).service

.PHONY: build install-service start stop restart status logs uninstall

build:
	go build -o cuba .

install-service: build
	@echo "Instalando servicio $(SERVICE_NAME)..."
	sudo bash -c 'cat > $(SERVICE_FILE)' <<EOF
[Unit]
Description=Cuba Service
After=network.target

[Service]
Type=simple
WorkingDirectory=$(shell pwd)
ExecStart=$(shell pwd)/cuba
Restart=always
RestartSec=3
User=root
Environment=PORT=1634

[Install]
WantedBy=multi-user.target
EOF

	sudo systemctl daemon-reload
	sudo systemctl enable $(SERVICE_NAME)

start:
	sudo systemctl start $(SERVICE_NAME)

stop:
	sudo systemctl stop $(SERVICE_NAME)

restart:
	sudo systemctl restart $(SERVICE_NAME)

status:
	sudo systemctl status $(SERVICE_NAME)

logs:
	sudo journalctl -u $(SERVICE_NAME) -f

uninstall:
	sudo systemctl stop $(SERVICE_NAME)
	sudo systemctl disable $(SERVICE_NAME)
	sudo rm -f $(SERVICE_FILE)
	sudo systemctl daemon-reload
