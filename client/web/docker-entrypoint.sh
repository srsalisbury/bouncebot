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

# Rewrite root-relative paths in built assets if BASE_PATH is set
if [ "${BASE_PATH:-/}" != "/" ]; then
  HTML_DIR=/usr/share/nginx/html

  # index.html: href="/..." and src="/..."
  sed -i "s|href=\"/|href=\"${BASE_PATH}|g; s|src=\"/|src=\"${BASE_PATH}|g" $HTML_DIR/index.html

  # JS bundle: "/filename.svg" and "/help/" references
  for f in $HTML_DIR/assets/*.js; do
    sed -i "s|\"/favicon_|\"${BASE_PATH}favicon_|g; s|\"/gear\.|\"${BASE_PATH}gear.|g; s|\"/timer\.|\"${BASE_PATH}timer.|g; s|\"/logo_|\"${BASE_PATH}logo_|g; s|\"/name_|\"${BASE_PATH}name_|g; s|\"/help/|\"${BASE_PATH}help/|g" "$f"
  done

  # CSS bundle: url(/fonts/...) and url(/pattern_...)
  for f in $HTML_DIR/assets/*.css; do
    sed -i "s|url(/fonts/|url(${BASE_PATH}fonts/|g; s|url(/pattern_|url(${BASE_PATH}pattern_|g" "$f"
  done

  # manifest.json
  if [ -f "$HTML_DIR/manifest.json" ]; then
    sed -i "s|\"src\": \"/|\"src\": \"${BASE_PATH}|g; s|\"start_url\": \"/|\"start_url\": \"${BASE_PATH}|g" $HTML_DIR/manifest.json
  fi

  echo "Rewrote asset paths with BASE_PATH=${BASE_PATH}"
fi

# Execute the CMD (nginx)
exec "$@"
