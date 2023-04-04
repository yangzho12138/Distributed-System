NODE_NUMBER=node1
FREQUENCY=0.5
build:
	go build
	python3 -u gentx.py $(FREQUENCY) | ./mp1_node $(NODE_NUMBER) config.txt