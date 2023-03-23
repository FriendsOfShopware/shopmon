#!/usr/bin/env sh

# import database if it doesnt exist yet

if [ ! -f /app/data/db/LOCAL_BINDING.sqlite3 ]; then
  echo "Importing database..."
  cat /app/db.sql | sqlite3 /app/data/db/LOCAL_BINDING.sqlite3
fi

nginx -g "daemon off;"  & # start nginx in background
P1=$!

if [ ! -f /app/data/.secret ]; then
  NEW_SECRET=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)
  echo $NEW_SECRET > /app/data/.secret
fi

# generate random APP_SECRET if not set
APP_SECRET=${APP_SECRET:-$(cat /app/data/.secret)}

cd /app/api && miniflare --port 5789 \
          -m \
          --kv kvStorage \
          --d1 __D1_BETA__LOCAL_BINDING \
          --d1-persist /app/data/db \
          --do SHOPS_SCRAPE=ShopScrape \
          --do PAGESPEED_SCRAPE=PagespeedScrape \
          --do USER_SOCKET=UserSocket \
          --do-persist /app/data/do \
          --binding USE_LOCAL_DATABASE=true \
          --binding APP_SECRET=$APP_SECRET \
          --binding PAGESPEED_API_KEY=$PAGESPEED_API_KEY \
          --binding MAILGUN_KEY=$MAILGUN_KEY \
          --binding MAILGUN_DOMAIN=$MAILGUN_DOMAIN \
          --build-command "node_modules/.bin/esbuild --bundle --outfile=worker.api.js --format=esm --conditions=node --minify --sourcemap src/index.ts --source-root=/" \
        worker.api.js &
P2=$!
wait $P1 $P2
