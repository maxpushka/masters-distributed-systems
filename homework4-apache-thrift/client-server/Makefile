.PHONY: server
server: server/gen-go
	cd server && go run main.go

.PHONY: client
client: client/gen-py client/venv
	cd client && . venv/bin/activate && python3 main.py

client/venv:
	cd client \
		&& python3 -m venv venv \
		&& . venv/bin/activate \
		&& pip install -r requirements.txt

server/gen-go:
	thrift --gen go -o server calculator.thrift

client/gen-py:
	thrift --gen py -o client calculator.thrift

clean:
	rm -rf server/gen-go client/gen-py client/venv
