#!/bin/sh
# Generate runtime config from environment variables
# This runs at container startup, before nginx starts

CONFIG_FILE=/usr/share/nginx/html/config.js

cat <<EOF > $CONFIG_FILE
// Runtime configuration - generated at container startup
window.APP_CONFIG = {
  API_BASE_URL: "${API_BASE_URL:-/api}",
  BASE_PATH: "${BASE_PATH:-/}"
};
EOF

# Make readable by nginx
chmod 644 $CONFIG_FILE

echo "Generated $CONFIG_FILE with API_BASE_URL=${API_BASE_URL:-/api} BASE_PATH=${BASE_PATH:-/}"

# Rewrite paths in manifest.json if BASE_PATH is set
MANIFEST_FILE=/usr/share/nginx/html/manifest.json
if [ "${BASE_PATH:-/}" != "/" ] && [ -f "$MANIFEST_FILE" ]; then
  sed -i "s|\"src\": \"/|\"src\": \"${BASE_PATH}|g; s|\"start_url\": \"/|\"start_url\": \"${BASE_PATH}|g" $MANIFEST_FILE
  echo "Rewrote paths in manifest.json with BASE_PATH=${BASE_PATH}"
fi

# Execute the CMD (nginx)
exec "$@"
