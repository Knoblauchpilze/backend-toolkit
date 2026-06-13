package db

import (
	"encoding/json"
	"errors"
	"fmt"

	berrors "github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
)

type DatabaseError struct {
	Message    string
	SqlCode    string
	Schema     string
	Table      string
	Column     string
	Constraint string
	Cause      error
}

func AsDatabaseError(err error) (*DatabaseError, bool) {
	var dbErr *DatabaseError
	if errors.As(err, &dbErr) {
		return dbErr, true
	}

	return nil, false
}

func (e *DatabaseError) Error() string {
	out := fmt.Sprintf("%s, code: %s", e.Message, e.SqlCode)

	if e.Cause != nil {
		out += fmt.Sprintf(" (cause: %v)", e.Cause.Error())
	}

	return out
}

func (e *DatabaseError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Code       berrors.ErrorCode `json:"code"`
		SqlCode    string            `json:"sql_code"`
		Message    string            `json:"message,omitempty"`
		Schema     string            `json:"schema"`
		Table      string            `json:"table"`
		Column     string            `json:"column"`
		Constraint string            `json:"constraint"`
		Cause      json.RawMessage   `json:"cause,omitempty"`
	}{
		Code:       errSqlViolation,
		SqlCode:    e.SqlCode,
		Message:    e.Message,
		Schema:     e.Schema,
		Table:      e.Table,
		Column:     e.Column,
		Constraint: e.Constraint,
		Cause:      e.marshalCause(),
	})
}

func (e *DatabaseError) marshalCause() json.RawMessage {
	if e.Cause == nil {
		return nil
	}

	var out []byte

	// Voluntarily ignoring the marshalling errors as there's nothing we
	// can do about it.
	if impl, ok := e.Cause.(*DatabaseError); ok {
		out, _ = json.Marshal(impl)
	} else if impl, ok := e.Cause.(*berrors.ErrorWithCode); ok {
		out, _ = json.Marshal(impl)
	} else {
		out, _ = json.Marshal(e.Cause.Error())
	}

	return out
}
