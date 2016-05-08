package lib

import "encoding/json"

func jsonToArray(data []byte) ([]interface{}, error) {
	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err != nil {
		logger.Errorf("Failed to decode JSON from pubnub: %v", err)
		return nil, err
	}

	return arr, nil
}
