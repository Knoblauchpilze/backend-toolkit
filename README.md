# backend-toolkit

# Overview

This project uses the following technologies:

- [postgres](https://www.postgresql.org/) for the databases.
- [go](https://go.dev/) as the server backend language.

# Badges

[![Build and test packages](https://github.com/Knoblauchpilze/backend-toolkit/actions/workflows/build-and-test-packages.yml/badge.svg)](https://github.com/Knoblauchpilze/backend-toolkit/actions/workflows/build-and-test-packages.yml)

[![codecov](https://codecov.io/gh/Knoblauchpilze/backend-toolkit/graph/badge.svg?token=GDVROJ3V4Q)](https://codecov.io/gh/Knoblauchpilze/backend-toolkit)

# Why this project

This project was born from creating multiple go projects for backend services and realizing that we mostly need the same starter-pack in regards to finding a web framework, building a CI, interacting with a database and so on.

Most of these components are very similar from one project to the next and benefit greatly from being shared. This way all bugs can be easily ported to all projects through a versioning system and we can also very easily import it in a new project.

Additionally we don't need to worry so much about testing those common packages as it can be done directly in the base module.

# Sample code

A bit of code is usually really helpful to see what's possible with a certain module. So here's what possible with this project:

```go
package main

import (
	"context"
	"net/http"
	"os"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/db"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/db/postgresql"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/logger"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/rest"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/server"
	"github.com/labstack/echo/v4"
)

func main() {
	// Create a logger printing to standard output
	log := logger.New(logger.NewPrettyWriter(os.Stdout))

	// Create the connection to access the database
	dbConfig := postgresql.NewConfigForLocalhost("my-database", "my-user", "my-password")

	conn, err := db.New(context.Background(), dbConfig)
	if err != nil {
		log.Errorf("Failed to create db connection: %v", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	// Create the server
	serverConfig := server.Config{
		Port: 1234,
	}
	s := server.NewWithLogger(serverConfig, log)

	// Add a route with some handler
	route := rest.NewRoute(http.MethodGet, "/info", infoHandlerGenerator(conn))
	if err := s.AddRoute(route); err != nil {
		log.Errorf("Failed to register route %v: %v", route.Path(), err)
		os.Exit(1)
	}

	// Start the server
	err = s.Start(context.Background())
	if err != nil {
		log.Errorf("Error while serving: %v", err)
		os.Exit(1)
	}
}

func infoHandlerGenerator(conn db.Connection) echo.HandlerFunc {
	return func(c echo.Context) error {
		sqlQuery := "SELECT count FROM my-table"

		// Use the connection to query the database and unmarshal the result
		// easily in an integer or a struct or anything you want
		value, err := db.QueryOne[int](c.Request().Context(), conn, sqlQuery)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to query database")
		}

		return c.String(http.StatusOK, value)
	}
}
```

By using only packges provided in this repository we are able to setup a server handling requests and fetching information from a database very easily.

The key features of the project are:

- a simple way to configure a connection to a `postgres` database using [pgx](https://github.com/jackc/pgx).
- an easy to use server using [echo](https://echo.labstack.com/) as a base.
- a powerful logging system that leverages [zerolog](https://github.com/rs/zerolog) and integrates it with `echo`.

We define multiple tags and versions in this repository to make it easy to pinpoint a specific behavior and upgrade when needed.

# Key features

## Errors

A fundamental aspect of Go is that error should be integrated as part of the normal flow of a program. To this end, we usually create quite a lot of distinct errors and can need to wrap errors when they are interpreted by a higher layer of the programs we write.

In this project we define the concept of an [error with code](https://github.com/Knoblauchpilze/backend-toolkit/blob/master/pkg/errors/error.go#L15) as follows:

```go
type ErrorWithCode interface {
	Code() ErrorCode
}
```

The idea is that we can have an error with a code attached to it so that customers/developers can communicate specific problems that make it easy to identify afterwards what went wrong.

The package provides multiple convenience methods to create such an error either from scratch (meaning without previous error) or by wrapping an existing error (typically when a dependent third party library fails with an unrecoverable error that can't be handled at the current level).

The errors are also defining a message that can explain the code. An error can also be marshalled to JSON so that it's easy to interpret in HTTP calls. This looks like so:

```json
{
  "Code": 1,
  "Message": "hihi",
  "Cause": {
    "Code": 1,
    "Message": "haha"
  }
}
```

It is encouraged to create custom error codes specialized to our business logic.

## Logging

As a transverse concern, logging is usually quite important in a backend service. The main attributes we want to guarantee with a common package is:

- easy to read.
- easily customizable to allow display of headers and prefixes (typically modules or services).
- ability to correlate logs for a request with one another.

To this end we used some capabilities provided by `echo` and `zerolog` and tried to make them work in combination.

### Echo context

By default a handler using `echo` has the following prototype:

```go
type HandlerFunc func(c echo.Context) error
```

The `echo.Context` contains a logger which is attached by default to each request. It stems from the general logger configured when instantiating the `echo.Echo` object.

### Binding zerolog to echo logger

The `zerolog` package and the `echo` package have slightly different interfaces to allow logging. In the future we might want to use a different logging backend. Therefore it does not seem very secure to expose the internals of the logging system we use outside of the package.

In the [logger](pkg/logger) package we defined a general log interface (`logger.Logger`) and provide some adapters to convert to other types. This is a typical way to bind different logging systems (see e.g. [pgx's adapter](https://github.com/jackc/pgx-zerolog) for `zerolog`).

For this project we have the following function:

```go
func Wrap(log Logger) echo.Logger {
	/* ... */
}
```

This allows to create a logger as usual and pass it over to the `echo` object easily.

## Database interaction

An important part of a backend service is usually to interact with some database where the information is stored. For most of the projects we had to work with in the past this meant spinning up a `postgres` database and interact with it.

We've been using the [pgx](https://github.com/jackc/pgx) for a long time and found it quite versatile. The `db` package is using it under the hood but hiding some of the internals in an attempt to allow easily upgrading to newer version and hide some of the complexity of managing the connection to the database.

### The connection

The main type brought by the `db` package is the [db.Connection](pkg/db/connection.go). It allows to start a transaction and execute some SQL code.

### Querying

`pgx` defines two main concepts: `Exec` and `Query`. The difference is explained in [this StackOverflow](https://stackoverflow.com/questions/60180651/what-are-the-differences-between-queryrow-and-exec-in-golang-sql-package) post and boils down (roughly) to whether we use `SELECT` or some other statement.

`pgx` defines some very convenient methods to query data and automatically scan it into the fields of a struct or return it as a specific type. This eliminates a lot of the boilerplate we had to do in the past with the `rows.Scan(...)`.

In order to leverage generics, we would like to offer something like:

```go
type myStruct struct {
	A int
	B string
}

func foo(conn db.Connection) error {
	s, err := conn.Query[myStruct](conn, "SELECT A, B FROM my_table")
	if err != nil {
		return err
	}

	fmt.Printf("My struct: %v\n", s)

	return nil
}
```

The problem with this syntax is that the generic type to assign to the `Query` method would have to be known when creating the connection. Additionally we can't change the generic type of the `Query` method when a connection is created so it's not so flexible.

To this end, the `db` package defines a free function like below:

```go
func Query[T any](conn db.Connection, sqlQuery string, arguments ...any) (T, error) {
	/* ... */
}
```

## The rest server

Another common aspect of offering a backend service is to have an HTTP server. In the past we usually used the [echo](https://echo.labstack.com/) framework. Although it's already providing some good abstraction, we noticed that some operations were quite common:

- configuring the server (base path, port)
- start and stop gracefully
- register routes

This is the purpose of the [rest](pkg/rest) and [server](pkg/server) packages: they define utilities that can be used to easily register routes and attach them to a server. This server can in turn started and stopped easily.

## Middleware

No matter the project and what HTTP handlers are actually doing, it's common that we expect some processing to happen for all of them. Typical examples are:

- error recovery
- timing
- observability

In Go (and in most HTTP framework) those concerns are usually handled through middlewares. A middleware is a piece of code that 'decorates' an existing handler to enhance its capabilities. A typical example is a rate-limiting middleware which keeps track of how often an endpoint was called and by whom and denies some requests in case too many are received.

### Request tracing

An important aspect of microservices is tracing. This allows to effectively follow the path of a request across services boundaries and is usually accomplished by adding a _correlation id_ to a request.

In this toolkit we act on this premise with two middlewares, described below.

### The response envelope middleware

In order to provide consistent response format across an APIs, the [response_envelope](pkg/middleware/response_envelope.go) middleware captures the output of any incoming request and wraps it into something that looks like the following:

```json
{
  "requestId": "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1",
  "status": "SUCCESS",
  "details": {
    "value": 12
  }
}
```

This clearly indicates:

- the request identifier
- whether it succeeded or not
- potential details about the success or failure

Having something like this in place for an API allows consumers to easily know whether a request was successful or not. Additionally we can rely on HTTP status codes to provided information about what went wrong.

The `ResponseEnvelope` middleware is added by default to the `Server`.

### A note on generating the request identifier

In order to keep track of the journey of a request in the microservice architecture, the `ResponseEnvelope` middleware tries to retrieve an existing identifier from the headers of the request: if this exists, it uses it as a request id. If not, it generates a new one.

This allows to make sure that if the request is forwarded to another service (and as long as **the code is attaching the request id to the new HTTP request**) we will also see traces for the other service with this request's id.

### Logging for a request

Any backend service usually produces logs: in case a request fails or misbehave, it might be interesting to inspect traces for this specific request.

We already mentioned in the [logging](#logging) section that this framework makes it easy to configure a global logger than can be shared across the controllers and services.

Additionally with the [RequestTracer](pkg/middleware/request_tracer.go) middleware we can attach a custom logger with a prefix matching the request identifier.

This looks like the following (for an example with the [user-service](https://github.com/Knoblauchpilze/user-service)):

```
2024-12-15 13:55:26 INF [95ad22c8-8854-466c-83f7-2630fac365ba] GET localhost:60001/v1/users processed in 14.187412ms -> 200
```

We clearly see which request it is and can correlate the request across multiple services.

# Installation

The tools described below are directly used by the project. It is mandatory to install them in order to build the project locally.

See the following links:

- [golang](https://go.dev/doc/install): this project was developed using go `1.23.2`.
- [golang migrate](https://github.com/golang-migrate/migrate/blob/master/cmd/migrate/README.md): following the instructions there should be enough.
- [postgresql](https://www.postgresql.org/) which can be taken from the packages with `sudo apt-get install postgresql-14` for example.

We also assume that this repository is cloned locally and available to use. To achieve this, just use the following command:

```bash
git clone git@github.com:Knoblauchpilze/backend-toolkit.git
```

**Note:** the migrate tool and postgres are only used for the test database and for tests. If you don't install them it means you'll not be able to run the tests locally.

# How to extend this project

As this project is meant to contain common tools to kick-start a backend project in Go, additions are welcomed. As a general rule, anything we add here should be:

- generic enough so that it can be used in multiple projects.
- well tested in order to guarantee a good level of quality for the base code.
- convenient to use either with structures from the standard library or from this project.

If you think what you want to propose fits those criteria, you can open a [new PR](https://github.com/Knoblauchpilze/backend-toolkit/pulls) in this project or [an issue](https://github.com/Knoblauchpilze/backend-toolkit/issues).

# Create a new release

A convenience script is provided in the [scripts](scripts) folder to create a new release: [create_release.sh](scripts/create_release.sh).

You can run this script as follows:

```bash
./create_release.sh v1.2.3
```

The version is optional: in case it's not provided, the script will try to determine the latest one and add one to it. Typically if the latest published version available in the repository (**locally**) is `v1.2.3`, running the script without arguments will result in a version `v1.2.3.1` being created.

The new version will automatically be published to the public repository. Therefore use this script with care.
