BIN=strix
SCRIPT=static/js/bundle.js

build: $(SCRIPT) $(BIN)

$(SCRIPT): src/js/* src/css/*
	yarn install
	node ./node_modules/.bin/webpack --config ./webpack.config.js

$(BIN): *.go
	go build -v
