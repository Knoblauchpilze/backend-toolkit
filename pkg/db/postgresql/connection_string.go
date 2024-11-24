package postgresql

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

const connectionStringPrefix = "postgresql://"

// https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNECT-CONNECT-TIMEOUT
const connectTimeOutKey = "connect_timeout"

func generateConnectionString(config Config) string {
	// https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING-URIS
	user := generateUserSpec(config.User, config.Password)
	host := generateHostSpec(config.Host, config.Port)
	params := generateParamSpec((config.ConnectTimeout))

	out := connectionStringPrefix
	if user != "" {
		out += user
	}

	if host != "" {
		if user != "" {
			out += "@"
		}
		out += host
	}

	if config.Database != "" {
		out += fmt.Sprintf("/%s", config.Database)
	}

	if params != "" {
		out += fmt.Sprintf("?%s", params)
	}

	return out
}

func generateUserSpec(user string, password string) string {
	userSpec := fmt.Sprintf("%s:%s", url.QueryEscape(user), url.QueryEscape(password))
	return strings.TrimSuffix(userSpec, ":")
}

func generateHostSpec(host string, port uint16) string {
	hostSpec := fmt.Sprintf("%s:%d", url.QueryEscape(host), port)
	return strings.TrimSuffix(hostSpec, ":0")
}

func generateParamSpec(connectionTimeout time.Duration) string {
	if connectionTimeout == 0 {
		return ""
	}
	return fmt.Sprintf("%s=%d", connectTimeOutKey, int(connectionTimeout.Seconds()))
}
