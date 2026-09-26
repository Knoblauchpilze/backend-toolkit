package rest

import "encoding/json"

func decodeJSONOrString(data []byte) any {
	var out any
	err := json.Unmarshal(data, &out)
	if err == nil {
		return out
	}

	return string(data)
}
