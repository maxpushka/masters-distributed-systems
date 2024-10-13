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
$(DATA_FILE):
	go run generate_data/main.go --numRecords $(NUM_RECORDS) --numWorkers $(NUM_WORKERS)
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
.PHONY: test-clickhouse
test-clickhouse:
	cd test && go run . --db clickhouse

# Query PostgreSQL
.PHONY: test-postgres
test-postgres:
	cd test && go run . --db postgres

# Clean up containers and data
.PHONY: clean
clean: down
	rm -f $(DATA_FILE)
