#!/bin/sh
set -e

# Bind mounts from the host are often root-owned (especially when Docker creates
# the directories on first run). The app runs as appuser and must write here.
if [ "$(id -u)" = "0" ]; then
	for dir in /data/storage /data/disks; do
		if [ -d "$dir" ]; then
			chown -R appuser:appuser "$dir"
		fi
	done
	exec su-exec appuser /app/lcloud
fi

exec /app/lcloud
