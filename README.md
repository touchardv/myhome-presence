# myhome-presence

A web service for querying, managing and tracking family members presence at home.

Please note this is a work-in-progress; I use this project to:

* Continue learning the Go language.
* Experiment with presence detection at home (using IP ping, Bluetooth...).
* Complete my "home" automation/security/... software.
* Have fun.

## Building

Requirements:

* The `make` command (e.g. [GNU make](https://www.gnu.org/software/make/manual/make.html)).
* The [Golang toolchain](https://golang.org/doc/install) (version 1.16 or later).

In a shell, execute: `make` (or `make build`).

The build artifacts can be cleaned by using: `make clean`.

## Running

In a shell, execute: `make run`

## API

The API is documented using the [OpenAPI](https://swagger.io/specification/) specification.

The Swagger UI for consuming the API is reachable from http://localhost:8080.
Note: when using the Chrome web browser, in order to get the web UI to work, one should ensure that "Insecure content" permission is allowed (Swagger UI is served via https but here the API specification is server via http).
