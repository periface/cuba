#!/usr/bin/env bash

set -e

SERVICE_NAME="cuba"
BINARY_NAME="cuba"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
WORKDIR="$(pwd)"

echo "🔧 Construyendo binario..."
go build -o "${BINARY_NAME}" .

echo "📦 Instalando servicio systemd..."

sudo tee "${SERVICE_FILE}" > /dev/null <<EOF
[Unit]
Description=Cuba Service
After=network.target

[Service]
Type=simple
WorkingDirectory=${WORKDIR}
ExecStart=${WORKDIR}/${BINARY_NAME}
Restart=always
RestartSec=3
User=root
Environment=PORT=1634

[Install]
WantedBy=multi-user.target
EOF

echo "🔄 Recargando systemd..."
sudo systemctl daemon-reload

echo "✅ Habilitando servicio..."
sudo systemctl enable "${SERVICE_NAME}"

echo "🚀 Reiniciando servicio..."
sudo systemctl restart "${SERVICE_NAME}"

echo "📊 Estado del servicio:"
sudo systemctl status "${SERVICE_NAME}" --no-pager
