BINARY_NAME=node
FREQUENCY=2
NODE_NUMBER=node1
PORT_NUMBER=1234
build:
	go build
	python3 -u generator.py $(FREQUENCY) | ./$(BINARY_NAME) $(NODE_NUMBER) 172.22.158.205 $(PORT_NUMBER)
