package postgresql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnit_GenerateConnectionString(t *testing.T) {
	type testCase struct {
		host           string
		port           uint16
		database       string
		user           string
		password       string
		connectTimeout time.Duration

		expectedConnectionString string
	}

	testCases := []testCase{
		{
			expectedConnectionString: "postgresql://",
		},
		{
			database:                 "mydb",
			expectedConnectionString: "postgresql:///mydb",
		},
		{
			host:                     "localhost",
			expectedConnectionString: "postgresql://localhost",
		},
		{
			host:                     "179.52.21.34",
			expectedConnectionString: "postgresql://179.52.21.34",
		},
		{
			host:                     "http://1.2.3.4/my-db-host",
			expectedConnectionString: "postgresql://http%3A%2F%2F1.2.3.4%2Fmy-db-host",
		},
		{
			host:                     "localhost",
			port:                     5433,
			expectedConnectionString: "postgresql://localhost:5433",
		},
		{
			host:                     "localhost",
			database:                 "mydb",
			expectedConnectionString: "postgresql://localhost/mydb",
		},
		{
			host:                     "localhost",
			user:                     "user",
			expectedConnectionString: "postgresql://user@localhost",
		},
		{
			host:                     "localhost",
			user:                     "my complex#user",
			expectedConnectionString: "postgresql://my+complex%23user@localhost",
		},
		{
			host:                     "localhost",
			password:                 "?@ dhard-password#",
			expectedConnectionString: "postgresql://:%3F%40+dhard-password%23@localhost",
		},
		{
			host:                     "localhost",
			user:                     "user",
			password:                 "secret",
			expectedConnectionString: "postgresql://user:secret@localhost",
		},
		{
			host:                     "localhost",
			user:                     "other",
			database:                 "otherdb",
			connectTimeout:           10 * time.Second,
			expectedConnectionString: "postgresql://other@localhost/otherdb?connect_timeout=10",
		},
		{
			connectTimeout:           10 * time.Second,
			expectedConnectionString: "postgresql://?connect_timeout=10",
		},
	}

	for _, testCase := range testCases {
		t.Run("", func(t *testing.T) {
			config := Config{
				Host:           testCase.host,
				Port:           testCase.port,
				Database:       testCase.database,
				User:           testCase.user,
				Password:       testCase.password,
				ConnectTimeout: testCase.connectTimeout,
			}

			actual := generateConnectionString(config)

			assert.Equal(t, testCase.expectedConnectionString, actual)
		})
	}
}
