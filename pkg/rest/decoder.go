package rest

import "encoding/json"

func DecodeJSONOrString(data []byte) (any, error) {
	var out any
	err := json.Unmarshal(data, &out)
	if err == nil {
		return out, nil
	}

	return string(data), nil
}
