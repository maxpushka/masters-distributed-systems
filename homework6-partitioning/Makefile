NAME = postgres
# PSQL = docker exec -it $(NAME) psql --dbname='postgres://postgres:postgres@localhost:5432'
PSQL = psql --dbname='postgres://postgres:postgres@localhost:5432'

up:
	docker run --rm -e POSTGRES_PASSWORD=postgres -e POSTGRES_HOST_AUTH_METHOD=trust -p 5432:5432 --name $(NAME) postgres:16

migrate:
	$(PSQL) -c "CREATE TABLE measurements ( city_id int not null, logdate date not null, peaktemp int, unitsales int ) PARTITION BY RANGE (logdate);"
	$(PSQL) -c "CREATE TABLE measurements_old PARTITION OF measurements FOR VALUES FROM ('2000-01-01') TO ('2024-09-01');"
	$(PSQL) -c "CREATE TABLE measurements_new PARTITION OF measurements FOR VALUES FROM ('2024-09-01') TO ('2030-01-01');"

insert:
	$(PSQL) -c "INSERT INTO measurements (city_id, logdate, peaktemp, unitsales) VALUES (1, '2020-10-05', 42, 120), (2, '2024-10-05', 33, 210);"

test:
	$(PSQL) -c "SELECT * FROM measurements;"
	$(PSQL) -c "SELECT * FROM measurements_old;"
	$(PSQL) -c "SELECT * FROM measurements_new;"
