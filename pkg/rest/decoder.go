package rest

import "encoding/json"

func DecodeJSONTo[T any](data []byte) (T, error) {
	var out T
	err := json.Unmarshal(data, &out)
	if err != nil {
		return out, err
	}

	return out, nil
}

func DecodeRawBytes(data []byte) ([]byte, error) {
	return data, nil
}

func DecodeString(data []byte) (string, error) {
	return string(data), nil
}

func DecodeJSONOrString(data []byte) (any, error) {
	var out any
	err := json.Unmarshal(data, &out)
	if err == nil {
		return out, nil
	}

	return string(data), nil
}
