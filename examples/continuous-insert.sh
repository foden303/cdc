#!/usr/bin/env bash
# Bulk insert users & products only (no orders) for faster CDC benchmark
# Usage: ./continuous-insert.sh [USERS] [PRODUCTS] [ITERATIONS] [CONCURRENCY]
# Defaults: USERS=10000, PRODUCTS=1000, ITERATIONS=0(infinite), CONCURRENCY=1

set -euo pipefail

USERS=${1:-10000}
PRODUCTS=${2:-1000}
ITER=${3:-0}
CONCURRENCY=${4:-1}

CONTAINER="cdc-source-db"
DB_USER="cdc_user"
DB_NAME="source_db"

# Generate N CSV lines for users and pipe to psql COPY (use temp table + INSERT ... ON CONFLICT)
bulk_insert_users() {
  local n=$1
  awk -v n="$n" -v seed="$RANDOM" 'BEGIN{srand(seed); for(i=1;i<=n;i++){ id=int(rand()*1e9); status=(rand()>0.2?"active":"inactive"); printf "user_%d,user_%d@example.com,User %d,%s\n", id, id, id, status}} END{print "\\."}' \
    | docker exec -i "$CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 -c "CREATE TEMP TABLE tmp_users (LIKE users INCLUDING DEFAULTS); COPY tmp_users (username,email,full_name,status) FROM STDIN WITH (FORMAT csv); INSERT INTO users (username,email,full_name,status) SELECT username,email,full_name,status FROM tmp_users ON CONFLICT (email) DO NOTHING;" >/dev/null
}

# Generate N CSV lines for products and pipe to psql COPY (use temp table + INSERT ... ON CONFLICT)
bulk_insert_products() {
  local n=$1
  awk -v n="$n" -v seed="$RANDOM" 'BEGIN{srand(seed); cats[1]="electronics"; cats[2]="clothing"; cats[3]="food"; cats[4]="books"; cats[5]="sports"; cats[6]="home"; cats[7]="toys"; for(i=1;i<=n;i++){ id=int(rand()*1e9); cat=cats[int(rand()*7)+1]; price=sprintf("%.2f", (rand()*100)+1); stock=int(rand()*500); printf "Product %d - %s,SKU-%d,%s,%d,%s\n", id, cat, id, price, stock, cat}} END{print "\\."}' \
    | docker exec -i "$CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 -c "CREATE TEMP TABLE tmp_products (LIKE products INCLUDING DEFAULTS); COPY tmp_products (name,sku,price,stock,category) FROM STDIN WITH (FORMAT csv); INSERT INTO products (name,sku,price,stock,category) SELECT name,sku,price,stock,category FROM tmp_products ON CONFLICT (sku) DO NOTHING;" >/dev/null
}

worker() {
  local users="$USERS" products="$PRODUCTS" iter_limit="$ITER"
  local round=0
  while true; do
    round=$((round+1))
    printf "[worker %d pid %d] round %d: users=%s products=%s\n" "$1" "$$" "$round" "$users" "$products"

    bulk_insert_users "$users"
    bulk_insert_products "$products"

    printf "[worker %d pid %d] round %d done\n" "$1" "$$" "$round"

    if [ "$iter_limit" -ne 0 ] && [ "$round" -ge "$iter_limit" ]; then
      break
    fi
  done
}

# trap to kill background workers on Ctrl-C
trap 'echo "Stopping..."; kill 0; exit' INT TERM

for i in $(seq 1 "$CONCURRENCY"); do
  worker "$i" &
done
wait
