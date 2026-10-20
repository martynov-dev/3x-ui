#!/bin/sh
set -eu
if [ ! -f /etc/x-ui/x-ui.db ]; then
    /app/x-ui setting -username dev-admin -password LocalTest-3xui-2026 -port 2053 -webBasePath /
fi
mkdir -p /var/www/3x-ui-site
if [ ! -f /var/www/3x-ui-site/index.html ]; then
    printf '%s\n' '<h1>Fork backup test website</h1>' > /var/www/3x-ui-site/index.html
fi
cat > /etc/nginx/nginx.conf <<'EOF'
events {}
http {
include /etc/nginx/mime.types;
include /etc/nginx/conf.d/*.conf;
include /etc/nginx/http.d/*.conf;
}
EOF
cat > /etc/nginx/http.d/3x-ui-test.conf <<'EOF'
server {
    listen 80 default_server;
    server_name _;
    root /var/www/3x-ui-site;
    index index.html;
}
EOF
rm -f /etc/nginx/http.d/default.conf
nginx
exec /app/x-ui
