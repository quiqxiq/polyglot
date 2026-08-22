#!/usr/bin/env bash
set -e

echo "🚀 [1/4] Pulling latest changes from Git..."
cd "$(dirname "$0")/.."
git pull

echo "⚙️ [2/4] Building Go Backend Server..."
go build -v -o bin/polyglot-server ./cmd/server

echo "🎨 [3/4] Building Web Frontend..."
if command -v pnpm &> /dev/null; then
  pnpm --dir web run build
elif command -v npm &> /dev/null; then
  npm --prefix web run build
fi

echo "🔄 [4/4] Restarting Polyglot Systemd Service..."
if command -v sudo &> /dev/null; then
  sudo systemctl restart polyglot.service
  sudo systemctl status polyglot.service --no-pager
else
  systemctl restart polyglot.service
  systemctl status polyglot.service --no-pager
fi

echo "✅ Polyglot update completed successfully!"
