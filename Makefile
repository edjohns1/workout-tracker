default:
	go build -C ./src -o ../bin/out

run:
	go build -C ./src -o ../bin/out
	bin/out
