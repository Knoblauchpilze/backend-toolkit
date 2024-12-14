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

	serverConfig := server.Config{
		Port: 1234,
	}
	s := server.NewWithLogger(serverConfig, log)

	route := rest.NewRoute(http.MethodGet, "/info", infoHandlerGenerator(conn))
	if err := s.AddRoute(route); err != nil {
		log.Errorf("Failed to register route %v: %v", route.Path(), err)
		os.Exit(1)
	}

	err = s.Start(context.Background())
	if err != nil {
		log.Errorf("Error while serving: %v", err)
		os.Exit(1)
	}
}

func infoHandlerGenerator(conn db.Connection) echo.HandlerFunc {
	return func(c echo.Context) error {
		sqlQuery := "SELECT count FROM my-table"

		value, err := db.QueryOne[int](c.Request().Context(), conn, sqlQuery)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to query database")
		}

		return c.JSON(http.StatusOK, value)
	}
}
```

By using only packges provided by this package we are able to setup a server handling requests and fetching information from a database very easily.

The key features of the project are:

- a simple way to configure a connection to a `postgres` database.
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

## Database interaction

## The rest server

## Middleware

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
