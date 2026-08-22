#!/usr/bin/env bash
set -e

echo "🚀 [1/4] Pulling latest changes from Git..."
cd "$(dirname "$0")/.."
git pull

echo "⚙️ [2/4] Building Go Backend Server..."
go build -v -o bin/polyglot-server ./cmd/server

echo "🎨 [3/4] Building Web Frontend..."
if [ -d "web" ]; then
  if command -v pnpm &> /dev/null; then
    if [ ! -d "web/node_modules" ]; then
      pnpm --dir web install
    fi
    pnpm --dir web run build || echo "⚠️ Frontend build failed, using existing dist"
  elif command -v npm &> /dev/null; then
    if [ ! -d "web/node_modules" ]; then
      npm --prefix web install
    fi
    npm --prefix web run build || echo "⚠️ Frontend build failed, using existing dist"
  fi
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
