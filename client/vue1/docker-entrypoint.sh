#!/bin/sh
# Generate runtime config from environment variables
# This runs at container startup, before nginx starts

CONFIG_FILE=/usr/share/nginx/html/config.js

cat <<EOF > $CONFIG_FILE
// Runtime configuration - generated at container startup
window.APP_CONFIG = {
  API_BASE_URL: "${API_BASE_URL:-/api}"
};
EOF

# Make readable by nginx
chmod 644 $CONFIG_FILE

echo "Generated $CONFIG_FILE with API_BASE_URL=${API_BASE_URL:-/api}"

# Execute the CMD (nginx)
exec "$@"
