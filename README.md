# Strix

User interface for [Minerva](https://github.com/m-mizutani/minerva).

![strix-sample](https://user-images.githubusercontent.com/605953/73526104-d2554780-4453-11ea-8480-ec331638569f.png)

## Prerequisite

- Tools
  - go >= 1.13
  - yarn >= 1.21.1
- Resources
  - Endpoint URL of deployed [Minerva](https://github.com/m-mizutani/minerva/blob/master/README.md)

## Build

```sh
$ git clone https://github.com/m-mizutani/strix.git
$ cd strix
$ yarn install
$ node ./node_modules/.bin/webpack --optimize-minimize --config ./webpack.config.js
$ go build -v
```

## Deploy & Run

```sh
$ cp -r ./strix ./static /path/to/deployment
$ cd /path/to/deployment
$ ./strix -a 0.0.0.0 -p 8080 https://xxxxxxxx.execute-api.ap-northeast-1.amazonaws.com/prod
```

Then, open http://localhost:8080 if you run strix on your local PC.

## License

MIT License
