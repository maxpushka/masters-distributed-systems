from flask import Flask, request, jsonify
import pika
import json
import time

app = Flask(__name__)

# Налаштування RabbitMQ
rabbitmq_host = 'localhost'
connection = pika.BlockingConnection(pika.ConnectionParameters(host=rabbitmq_host))
channel = connection.channel()

# Створюємо чергу
channel.queue_declare(queue='input_queue')
channel.queue_declare(queue='output_queue')


@app.route('/calculate', methods=['POST'])
def calculate():
    data = request.json
    num1 = data['num1']
    num2 = data['num2']

    # Кладемо числа у RabbitMQ
    message = json.dumps({'num1': num1, 'num2': num2})
    channel.basic_publish(exchange='', routing_key='input_queue', body=message)

    # Чекаємо на відповідь з вихідної черги
    while True:
        method_frame, header_frame, body = channel.basic_get(queue='output_queue', auto_ack=True)
        if body:
            result = json.loads(body)
            return jsonify({'result': result['result']})
        time.sleep(0.1)  # Невелика пауза між спробами для зменшення навантаження на CPU


if __name__ == '__main__':
    app.run(debug=True, port=8080)
