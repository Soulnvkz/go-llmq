proto:
	protoc --go_out=. --go_opt=paths=source_relative \
    	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
    	./proto/dbee.proto

dbee:
	cp ./dbee/.env build/dbee/.env 
	rm -rf build/dbee/migrations/
	cp -r ./dbee/migrations/ build/dbee/
	cd ./dbee && go build -o ../build/dbee/

llm: 
	cd ./llm && go build -o ../build/llm/llm

server:
	cd ./server && go build -gcflags=-m -o ../build/server/server