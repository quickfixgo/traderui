# UI Trader Example Client

## Installation
To install the QuickFIX/Go example trading client, use `go get`:

```sh
$ go get github.com/quickfixgo/traderui
```

### Staying up to date
To update the example trading client to the latest version, use `go get -u github.com/quickfixgo/traderui`.

## Building the Client
```sh
make build
```

## Running the Client
```sh
./bin/traderui
```
This will try to connect to a FIX acceptor on `localhost:5001` and expose the UI on `localhost:8080`.

## Licensing
This software is available under the QuickFIX Software License. Please see the [LICENSE](https://github.com/quickfixgo/traderui/blob/main/LICENSE) for the terms specified by the QuickFIX Software License.
