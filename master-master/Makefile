all: up setup test

.PHONY: up
up:
	docker-compose up -Vd
	sleep 10

.PHONY: setup
setup:
	# Initialize the Nodes to create replication users as well as check Master Log / position and verify server_id
	docker exec -ti mysql1 sh -c "mysql -uroot --password=mysql1pass -e 'SOURCE /backup/initdb.sql;'"
	docker exec -ti mysql2 sh -c "mysql -uroot --password=mysql2pass -e 'SOURCE /backup/initdb.sql;'"
	# Setup the Replication source for both nodes
	docker exec -ti mysql2 sh -c "mysql -uroot --password=mysql2pass -e \"stop slave; CHANGE MASTER TO MASTER_HOST = 'mysql1', MASTER_USER = 'replicator', MASTER_PASSWORD = 'repl1234or', MASTER_LOG_FILE = 'mysql-bin.000003', MASTER_LOG_POS = 154; start slave; show slave status;\""
	docker exec -ti mysql1 sh -c "mysql -uroot --password=mysql1pass -e \"stop slave; CHANGE MASTER TO MASTER_HOST = 'mysql2', MASTER_USER = 'replicator', MASTER_PASSWORD = 'repl1234or', MASTER_LOG_FILE = 'mysql-bin.000003', MASTER_LOG_POS = 154; start slave; show slave status;\""

.PHONY: test
test:
	docker exec -ti mysql1 sh -c "mysql -uroot --password=mysql1pass -e 'use mydata; create table students (id int, name varchar(20));'"
	docker exec -ti mysql2 sh -c "mysql -uroot --password=mysql2pass -e 'use mydata; show tables; drop table students;'"
	docker exec -ti mysql1 sh -c "mysql -uroot --password=mysql1pass -e 'use mydata; show tables;'"
