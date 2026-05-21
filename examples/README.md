# Examples

Thư mục này chứa các ví dụ và scripts hỗ trợ test CDC pipeline.

## Cấu trúc

```
examples/
├── docker-compose.test.yaml   # Docker Compose cho test environment (NATS + PostgreSQL source/sink + Prometheus)
├── init-source.sql            # SQL khởi tạo source database (tables: users, products, orders)
├── init-sink.sql              # SQL khởi tạo sink database
├── seed-data.sh               # Script seed dữ liệu mẫu vào source database
└── README.md
```

## Sử dụng

### 1. Khởi động test environment

```bash
cd examples
docker compose -f docker-compose.test.yaml up -d
```

Sẽ khởi động:
- **NATS** (port 4222) — message broker
- **Source PostgreSQL** (port 5433) — database nguồn với WAL logical replication
- **Sink PostgreSQL** (port 5434) — database đích
- **Prometheus** (port 9095) — monitoring

### 2. Seed dữ liệu mẫu

```bash
./seed-data.sh [NUM_USERS] [NUM_PRODUCTS] [NUM_ORDERS]
```

Mặc định: 50 users, 30 products, 100 orders.

```bash
# Seed với giá trị mặc định
./seed-data.sh

# Seed custom
./seed-data.sh 100 50 200
```

### 3. Dừng environment

```bash
cd examples
docker compose -f docker-compose.test.yaml down
```

## Lưu ý

- Script `seed-data.sh` sử dụng `docker exec` nên không cần cài `psql` trên máy local.
- Source database đã bật `wal_level=logical` để hỗ trợ CDC replication.
