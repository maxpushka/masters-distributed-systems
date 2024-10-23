queue:
	docker run --rm -d --hostname rabbitmq --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq

.PHONY: api
api: api/venv
	cd api && . venv/bin/activate && python3 main.py

.PHONY: worker
worker:
	cd worker && go run main.go

.PHONY: test
test:
	curl -X POST http://localhost:8080/calculate -H "Content-Type: application/json" -d '{"num1": 5, "num2": 7}'

# ------------------------------------------------------- #

api/venv:
	cd api \
		&& python3 -m venv venv \
		&& . venv/bin/activate \
		&& pip install -r requirements.txt

clean:
	rm -rf api/venv
