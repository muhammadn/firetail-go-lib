package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func getDefaultBatchCallback(options BatchLoggerOptions) func([][]byte) error {
	return func(batchBytes [][]byte) error {
		if options.LogApiUrl == "" {
			return errors.New("no log api url set")
		}
		reqBytes := []byte{}
		for _, entry := range batchBytes {
			reqBytes = append(reqBytes, entry...)
			reqBytes = append(reqBytes, '\n')
		}
		req, err := http.NewRequest("POST", options.LogApiUrl, bytes.NewBuffer(reqBytes))
		if err != nil {
			return err
		}
		req.Header.Set("x-ft-api-key", options.LogApiKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		if res["message"] != "success" {
			return errors.New(fmt.Sprintf("got err response from firetail api: %v", res))
		}
		return nil
	}
}
