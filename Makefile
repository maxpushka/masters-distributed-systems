gen:
	thrift --gen go -o server calculator.thrift
	thrift --gen py -o client calculator.thrift
