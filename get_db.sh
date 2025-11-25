#!/bin/sh
# Directory to store the DB
DB_DIR="."
mkdir -p "$DB_DIR"

# Download DB-IP Lite (City) - Updates monthly
# Use curl with -L to follow redirects
curl -L -o "$DB_DIR/dbip-city-lite.mmdb.gz" \
    "https://download.db-ip.com/free/dbip-city-lite-$(date +%Y-%m).mmdb.gz"

# If the monthly file isn't out yet, fallback to last month
if [ ! -s "$DB_DIR/dbip-city-lite.mmdb.gz" ]; then
    LAST_MONTH=$(date -d "last month" +%Y-%m)
    curl -L -o "$DB_DIR/dbip-city-lite.mmdb.gz" \
        "https://download.db-ip.com/free/dbip-city-lite-$LAST_MONTH.mmdb.gz"
fi

# Decompress
gunzip -f "$DB_DIR/dbip-city-lite.mmdb.gz"
