queue:
	docker run --rm -d --hostname rabbitmq --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq

.PHONY: worker
worker:
	cd worker && go run main.go

.PHONY: api
api: api/venv
	cd api && . venv/bin/activate && python3 main.py

# ------------------------------------------------------- #

api/venv:
	cd api \
		&& python3 -m venv venv \
		&& . venv/bin/activate \
		&& pip install -r requirements.txt

clean:
	rm -rf api/venv
