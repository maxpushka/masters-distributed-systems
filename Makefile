# Data generation params
DATA_FILE = purchase_data.csv
NUM_RECORDS = 1000000
NUM_WORKERS = 24

# Database settings
CLICKHOUSE_CLIENT = clickhouse-client --time
POSTGRES_CLIENT = time docker exec -i postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)
POSTGRES_USER = user
POSTGRES_PASSWORD = password
POSTGRES_DB = postgres
POSTGRES_HOST = localhost
POSTGRES_PORT = 5432

# Docker commands
.PHONY: up
up:
	docker-compose up -d
	sleep 10
	$(MAKE) load-clickhouse
	$(MAKE) load-postgres

.PHONY: down
down:
	docker-compose down --volumes

# Generate data
DATA_FILE = purchase_data.csv
generate-data: $(DATA_FILE)

# Target to regenerate the CSV if needed
$(DATA_FILE): generate_data.go
	go run generate_data.go --numRecords $(NUM_RECORDS) --numWorkers $(NUM_WORKERS)
	touch $(DATA_FILE)

# Track the number of records to force regeneration when NUM_RECORDS changes
$(NUM_RECORDS_FILE):
	echo $(NUM_RECORDS) > $(NUM_RECORDS_FILE)

# Load data into ClickHouse
.PHONY: load-clickhouse
load-clickhouse: generate-data
	$(CLICKHOUSE_CLIENT) --query="CREATE TABLE IF NOT EXISTS purchases (receipt_id UInt32, store String, product String, quantity UInt32, price Float32, purchase_date DateTime) ENGINE = MergeTree() ORDER BY (store, product, purchase_date);"
	$(CLICKHOUSE_CLIENT) --query="INSERT INTO purchases FORMAT CSV" < $(DATA_FILE)

# Load data into PostgreSQL
.PHONY: load-postgres
load-postgres: generate-data
	until $(POSTGRES_CLIENT) -P pager=off -c '\l'; do \
		echo >&2 "$(date +%Y%m%dt%H%M%S) Postgres is unavailable - sleeping"; \
		sleep 1; \
	done
	$(POSTGRES_CLIENT) -c "CREATE TABLE IF NOT EXISTS purchases (receipt_id INTEGER, store VARCHAR(50), product VARCHAR(50), quantity INTEGER, price FLOAT, purchase_date TIMESTAMP);"
	$(POSTGRES_CLIENT) -c "\COPY purchases(receipt_id, store, product, quantity, price, purchase_date) FROM STDIN WITH CSV HEADER" < $(DATA_FILE)

# Query ClickHouse
.PHONY: query-clickhouse
run-clickhouse-tests:
	# 1. Порахувати кількість проданого товару
	$(CLICKHOUSE_CLIENT) --query="SELECT SUM(quantity) AS total_quantity_sold FROM purchases;"
	# 2. Порахувати вартість проданого товару
	$(CLICKHOUSE_CLIENT) --query="SELECT SUM(quantity * price) AS total_value_sold FROM purchases;"
	# 3. Порахувати вартість проданого товару за період
	$(CLICKHOUSE_CLIENT) --query="SELECT SUM(quantity * price) AS total_value_sold FROM purchases WHERE purchase_date BETWEEN '2024-01-01' AND '2024-01-31';"
	# 4. Порахувати скільки було придбано товару А в магазині В за період С
	$(CLICKHOUSE_CLIENT) --query="SELECT SUM(quantity) AS total_quantity_sold FROM purchases WHERE product = 'Чай' AND store = 'Магазин B' AND purchase_date BETWEEN '2024-01-01' AND '2024-01-31';"
	# 5. Порахувати скільки було придбано товару А в усіх магазинах за період С
	$(CLICKHOUSE_CLIENT) --query="SELECT SUM(quantity) AS total_quantity_sold FROM purchases WHERE product = 'Чай' AND purchase_date BETWEEN '2024-01-01' AND '2024-01-31';"
	# 6. Порахувати сумарну виручку магазинів за період С
	$(CLICKHOUSE_CLIENT) --query="SELECT store, SUM(quantity * price) AS total_revenue FROM purchases WHERE purchase_date BETWEEN '2024-01-01' AND '2024-01-31' GROUP BY store;"
	# 7. Вивести топ 10 купівель товарів по два за період С
	$(CLICKHOUSE_CLIENT) --query="SELECT least(p1.product, p2.product) AS product_1, greatest(p1.product, p2.product) AS product_2, COUNT(*) AS count FROM purchases p1 JOIN purchases p2 ON p1.receipt_id = p2.receipt_id WHERE p1.product != p2.product AND p1.purchase_date BETWEEN '2024-01-01' AND '2024-01-31' GROUP BY product_1, product_2 ORDER BY count DESC LIMIT 10;"
	# 8. Вивести топ 10 купівель товарів по три за період С
	$(CLICKHOUSE_CLIENT) --query="SELECT least(p1.product, p2.product, p3.product) AS product_1, arraySort([p1.product, p2.product, p3.product])[2] AS product_2, greatest(p1.product, p2.product, p3.product) AS product_3, COUNT(*) AS count FROM purchases p1 JOIN purchases p2 ON p1.receipt_id = p2.receipt_id JOIN purchases p3 ON p1.receipt_id = p3.receipt_id WHERE p1.product < p2.product AND p2.product < p3.product AND p1.product != p3.product AND p1.purchase_date BETWEEN '2024-01-01' AND '2024-01-31' GROUP BY product_1, product_2, product_3 ORDER BY count DESC LIMIT 10;"
	# 9. Вивести топ 10 купівель товарів по чотири за період С
	$(CLICKHOUSE_CLIENT) --query="SELECT  least(p1.product, p2.product, p3.product, p4.product) AS product_1,  arraySort([p1.product, p2.product, p3.product, p4.product])[2] AS product_2,  arraySort([p1.product, p2.product, p3.product, p4.product])[3] AS product_3,  greatest(p1.product, p2.product, p3.product, p4.product) AS product_4,  COUNT(*) AS count  FROM purchases p1  JOIN purchases p2 ON p1.receipt_id = p2.receipt_id  JOIN purchases p3 ON p1.receipt_id = p3.receipt_id  JOIN purchases p4 ON p1.receipt_id = p4.receipt_id  WHERE p1.product < p2.product AND p2.product < p3.product AND p3.product < p4.product  AND p1.product != p4.product  AND p1.purchase_date BETWEEN '2024-01-01' AND '2024-01-31'  GROUP BY product_1, product_2, product_3, product_4  ORDER BY count DESC  LIMIT 10;"

# Query PostgreSQL
.PHONY: query-postgres
run-postgres-tests:
	# 1. Порахувати кількість проданого товару
	$(POSTGRES_CLIENT) -c "SELECT SUM(quantity) AS total_quantity_sold FROM purchases;"
	# 2. Порахувати вартість проданого товару
	$(POSTGRES_CLIENT) -c "SELECT SUM(quantity * price) AS total_value_sold FROM purchases;"
	# 3. Порахувати вартість проданого товару за період
	$(POSTGRES_CLIENT) -c "SELECT SUM(quantity * price) AS total_value_sold FROM purchases WHERE purchase_date BETWEEN '2024-01-01' AND '2024-01-31';"
	# 4. Порахувати скільки було придбано товару А в магазині В за період С
	$(POSTGRES_CLIENT) -c "SELECT SUM(quantity) AS total_quantity_sold FROM purchases WHERE product = 'Чай' AND store = 'Магазин B' AND purchase_date BETWEEN '2024-01-01' AND '2024-01-31';"
	# 5. Порахувати скільки було придбано товару А в усіх магазинах за період С
	$(POSTGRES_CLIENT) -c "SELECT SUM(quantity) AS total_quantity_sold FROM purchases WHERE product = 'Чай' AND purchase_date BETWEEN '2024-01-01' AND '2024-01-31';"
	# 6. Порахувати сумарну виручку магазинів за період С
	$(POSTGRES_CLIENT) -c "SELECT store, SUM(quantity * price) AS total_revenue FROM purchases WHERE purchase_date BETWEEN '2024-01-01' AND '2024-01-31' GROUP BY store;"
	# 7. Вивести топ 10 купівель товарів по два за період С
	$(POSTGRES_CLIENT) -c "SELECT LEAST(p1.product, p2.product) AS product_1, GREATEST(p1.product, p2.product) AS product_2, COUNT(*) AS count FROM purchases p1 JOIN purchases p2 ON p1.receipt_id = p2.receipt_id WHERE p1.product != p2.product AND p1.purchase_date BETWEEN '2024-01-01' AND '2024-01-31' GROUP BY product_1, product_2 ORDER BY count DESC LIMIT 10;"
	# 8. Вивести топ 10 купівель товарів по три за період С
	$(POSTGRES_CLIENT) -c "SELECT LEAST(p1.product, p2.product, p3.product) AS product_1, (SELECT product FROM (VALUES (p1.product), (p2.product), (p3.product)) AS t(product) ORDER BY product LIMIT 1 OFFSET 1) AS product_2, GREATEST(p1.product, p2.product, p3.product) AS product_3, COUNT(*) AS count FROM purchases p1 JOIN purchases p2 ON p1.receipt_id = p2.receipt_id AND p1.product < p2.product JOIN purchases p3 ON p1.receipt_id = p3.receipt_id AND p2.product < p3.product WHERE p1.purchase_date BETWEEN '2024-01-01' AND '2024-01-31' GROUP BY product_1, product_2, product_3 ORDER BY count DESC LIMIT 10;"
	# 9. Вивести топ 10 купівель товарів по чотири за період С
	$(POSTGRES_CLIENT) -c "SELECT  LEAST(p1.product, p2.product, p3.product, p4.product) AS product_1,  (SELECT product FROM (VALUES (p1.product), (p2.product), (p3.product), (p4.product)) AS t(product) ORDER BY product LIMIT 1 OFFSET 1) AS product_2, (SELECT product FROM (VALUES (p1.product), (p2.product), (p3.product), (p4.product)) AS t(product) ORDER BY product LIMIT 1 OFFSET 2) AS product_3, GREATEST(p1.product, p2.product, p3.product, p4.product) AS product_4,  COUNT(*) AS count  FROM purchases p1  JOIN purchases p2 ON p1.receipt_id = p2.receipt_id AND p1.product < p2.product  JOIN purchases p3 ON p1.receipt_id = p3.receipt_id AND p2.product < p3.product  JOIN purchases p4 ON p1.receipt_id = p4.receipt_id AND p3.product < p4.product  WHERE p1.purchase_date BETWEEN '2024-01-01' AND '2024-01-31'  GROUP BY product_1, product_2, product_3, product_4  ORDER BY count DESC  LIMIT 10;"

# Clean up containers and data
.PHONY: clean
clean: down
	rm -f $(DATA_FILE)
