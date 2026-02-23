package rest

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

var multiSlashRegex = regexp.MustCompile("[/]+")

func sanitizePath(route string) string {
	route = fmt.Sprintf("/%s", route)
	route = multiSlashRegex.ReplaceAllString(route, "/")
	route = strings.TrimSuffix(route, "/")

	if len(route) == 0 {
		return "/"
	}

	return route
}

func ConcatenateEndpoints(basePath string, path string) string {
	concatenated := fmt.Sprintf("/%s/%s", basePath, path)
	return sanitizePath(concatenated)
}

func MarshalNilToEmptySlice[T any](in []T) ([]byte, error) {
	toMarshal := make([]T, 0)
	if in != nil {
		toMarshal = in
	}

	return json.Marshal(toMarshal)
}

func FetchIdFromQueryParam(key string, c *echo.Context) (exists bool, id uuid.UUID, err error) {
	maybeId := c.QueryParam(key)
	exists = (maybeId != "")
	if maybeId == "" {
		return exists, uuid.UUID{}, nil
	}

	id, err = uuid.Parse(maybeId)
	return exists, id, err
}
