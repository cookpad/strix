BIN=strix
SCRIPT=static/js/bundle.js

build: $(SCRIPT) $(BIN)

run: $(SCRIPT) $(BIN)
	./strix

debug:
	node devel/stub.js &
	go run -v . -a 127.0.0.1 -p 9080 http://127.0.0.1:8000/ &
	npm run start:dev

local:
	go run -v . -a 127.0.0.1 -p 9080 http://127.0.0.1:8000/ &
	npm run start:dev

$(SCRIPT): javascript/*
	yarn install
	node ./node_modules/.bin/webpack --optimize-minimize --config ./webpack.config.js

$(BIN): *.go
	go build -v
