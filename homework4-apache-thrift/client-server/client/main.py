from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

import sys
sys.path.append('gen-py')
from calculator import Calculator

def main():
    # Створюємо сокет
    transport = TSocket.TSocket('localhost', 9090)
    transport = TTransport.TBufferedTransport(transport)
    protocol = TBinaryProtocol.TBinaryProtocol(transport)

    # Створюємо клієнта
    client = Calculator.Client(protocol)
    transport.open()

    # Викликаємо метод
    result = client.add(5, 7)
    print(f"5 + 7 = {result}")

    transport.close()

if __name__ == "__main__":
    main()
