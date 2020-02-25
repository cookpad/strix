# Strix

Web User Interface for [Minerva](https://github.com/m-mizutani/minerva).

![strix-sample](https://user-images.githubusercontent.com/605953/75205792-0bb17680-57b8-11ea-8e27-ead912af1194.png)

## Prerequisite

- Tools
  - go >= 1.13
  - yarn >= 1.21.1
- Resources
  - Endpoint URL of deployed [Minerva](https://github.com/m-mizutani/minerva/blob/master/README.md)
  - Google OAuth 2.0 Client setting (a JSON file saved as `oatuh.json` ).
    - Set callback redirect URL as `http://localhost:8080/auth/google/callback` if you run strix on localhost.
    - See [docs](https://developers.google.com/identity/protocols/OAuth2) for more details.

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
$ export API_KEY=YOUR_API_GATEWEAY_KEY
$ cat oauth.json | jq
{
  "web": {
    "client_id": "xxxxxxxxxxxxxxxxxxxxxxxxxx.apps.googleusercontent.com",
    "project_id": "your-project-id",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_secret": "XXXXXXXXXXXXXXXXXXXXXXXXXXX",
    "redirect_uris": [
      "http://localhost:8080/auth/google/callback"
    ]
  }
}
$ ./strix -a 0.0.0.0 -p 8080 -g oauth.json https://xxxxxxxx.execute-api.ap-northeast-1.amazonaws.com/prod
```

Then, open http://localhost:8080 if you run strix on your local PC.

## License

MIT License
