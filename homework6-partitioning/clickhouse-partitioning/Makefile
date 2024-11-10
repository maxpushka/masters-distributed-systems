CLIENT_1 = docker exec -it chnode1 clickhouse-client
CLIENT_2 = docker exec -it chnode2 clickhouse-client

.PHONY: all
all:
	docker-compose up -Vd
	sleep 5
	$(MAKE) sharding
	docker-compose down --volumes

.PHONY: sharding
sharding:
	# Create database and table
	$(CLIENT_1) -q 'SHOW CLUSTERS'
	$(CLIENT_1) -q 'CREATE DATABASE db1 ON CLUSTER cluster_2S_1R'
	$(CLIENT_1) -q 'CREATE TABLE db1.table1 ON CLUSTER cluster_2S_1R (`id` UInt64, `column1` String) ENGINE = MergeTree ORDER BY id'
	sleep 1

	# Insert values into the table
	$(CLIENT_1) -q "INSERT INTO db1.table1 (id, column1) VALUES (1, 'abc')"
	$(CLIENT_2) -q "INSERT INTO db1.table1 (id, column1) VALUES (2, 'def')"
	sleep 1

	# You will see only the row that was inserted into that table on that node 
	$(CLIENT_1) -q 'SELECT * FROM db1.table1;'
	$(CLIENT_2) -q 'SELECT * FROM db1.table1;'
	sleep 1

	# Create a distributed table to query both shards on both nodes.
	# In this example, the `rand()` function is set as the sharding key
	# so that it randomly distributes each insert.
	$(CLIENT_1) -q 'CREATE TABLE db1.table1_dist ON CLUSTER cluster_2S_1R (`id` UInt64, `column1` String) ENGINE = Distributed("cluster_2S_1R", "db1", "table1", rand())'
	sleep 1

	# Query the distributed table to see both rows
	$(CLIENT_2) -q 'SELECT * FROM db1.table1_dist;'

