# myhome-presence

A web service for querying, managing and tracking home devices presence at home.

## Building

Requirements:

* The `make` command (e.g. [GNU make](https://www.gnu.org/software/make/manual/make.html)).
* The [Golang toolchain](https://golang.org/doc/install) (version 1.22 or later).

In a shell, execute: `make` (or `make build`).

The build artifacts can be cleaned by using: `make clean`.

## Running

In a shell, execute `make run` or `make run-image` to run from a container.

## API

The API is documented using the [OpenAPI](https://swagger.io/specification/) specification.

The Swagger UI for consuming the API is reachable from http://localhost:8080.
Note: when using the Chrome web browser, in order to get the web UI to work, one should ensure that "Insecure content" permission is allowed (Swagger UI is served via https but here the API specification is server via http).
