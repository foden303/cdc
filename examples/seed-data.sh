#!/bin/bash
# ============================================
# SEED DATA SCRIPT (via docker exec, no local psql needed)
# Usage: ./scripts/seed-data.sh [NUM_USERS] [NUM_PRODUCTS] [NUM_ORDERS]
# Default: 50 users, 30 products, 100 orders
# ============================================

set -e

NUM_USERS=${1:-50}
NUM_PRODUCTS=${2:-30}
NUM_ORDERS=${3:-100}

CONTAINER="cdc-source-db"
DB_USER="cdc_user"
DB_NAME="source_db"

# Helper: run SQL via docker exec
run_sql() {
  docker exec -e PGPASSWORD=cdc_pass "$CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -q -c "$1"
}

echo "🌱 Seeding source database via docker exec..."
echo "   Users: $NUM_USERS | Products: $NUM_PRODUCTS | Orders: $NUM_ORDERS"
echo ""

# --- Seed Users ---
echo "📝 Inserting $NUM_USERS users..."
for i in $(seq 1 $NUM_USERS); do
    run_sql "INSERT INTO users (username, email, full_name, status) VALUES (
        'user_${i}',
        'user_${i}@example.com',
        'User Number ${i}',
        CASE WHEN random() > 0.2 THEN 'active' ELSE 'inactive' END
    ) ON CONFLICT (email) DO NOTHING;"
done
echo "   ✅ Users done"

# --- Seed Products ---
echo "📝 Inserting $NUM_PRODUCTS products..."
CATEGORIES=("electronics" "clothing" "food" "books" "sports" "home" "toys")
for i in $(seq 1 $NUM_PRODUCTS); do
    CAT_IDX=$((RANDOM % ${#CATEGORIES[@]}))
    CATEGORY=${CATEGORIES[$CAT_IDX]}
    PRICE=$(echo "scale=2; ($RANDOM % 10000) / 100 + 1" | bc)
    STOCK=$((RANDOM % 500 + 1))
    run_sql "INSERT INTO products (name, sku, price, stock, category) VALUES (
        'Product ${i} - ${CATEGORY}',
        'SKU-$(printf '%05d' $i)',
        ${PRICE},
        ${STOCK},
        '${CATEGORY}'
    ) ON CONFLICT (sku) DO NOTHING;"
done
echo "   ✅ Products done"

# --- Seed Orders ---
echo "📝 Inserting $NUM_ORDERS orders..."
for i in $(seq 1 $NUM_ORDERS); do
    USER_ID=$((RANDOM % NUM_USERS + 1))
    PRODUCT_ID=$((RANDOM % NUM_PRODUCTS + 1))
    QTY=$((RANDOM % 5 + 1))
    STATUSES=("pending" "confirmed" "shipped" "delivered" "cancelled")
    STATUS_IDX=$((RANDOM % ${#STATUSES[@]}))
    STATUS=${STATUSES[$STATUS_IDX]}
    TOTAL=$(echo "scale=2; ($RANDOM % 10000) / 100 * $QTY" | bc)
    run_sql "INSERT INTO orders (user_id, product_id, quantity, total_amount, status) VALUES (
        ${USER_ID},
        ${PRODUCT_ID},
        ${QTY},
        ${TOTAL},
        '${STATUS}'
    );"
done
echo "   ✅ Orders done"

echo ""
echo "🎉 Seed complete!"
echo ""
echo "📊 Verify counts:"
run_sql "SELECT 'users' as table_name, count(*) FROM users
     UNION ALL
     SELECT 'products', count(*) FROM products
     UNION ALL
     SELECT 'orders', count(*) FROM orders;"
