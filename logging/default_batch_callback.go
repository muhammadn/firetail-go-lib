package logging

import (
	"bytes"
	_ "encoding/json"
	_ "errors"
	"net/http"

	"log"
	"io/ioutil"
)

func getDefaultBatchCallback(options BatchLoggerOptions) func([][]byte) {
	sendBatch := func(batchBytes [][]byte) error {
                log.Println("SENDING BATCH")

                reqBytes := []byte{}
                for _, entry := range batchBytes {
                        reqBytes = append(reqBytes, entry...)
                        reqBytes = append(reqBytes, '\n')
                }

/*                log.Println("reqBODY: ", bytes.NewBuffer(reqBytes).String())
                log.Println("API URL: ", options.LogApiUrl)
                log.Println("API KEY: ", options.LogApiKey)

                reqq, err := http.NewRequest("POST", "https://api.logging.eu-west-1.prod.firetail.app/logs/bulk", bytes.NewBuffer(reqBytes))
                if err != nil {
                        return err
                }
                reqq.Header.Set("x-ft-api-key", "dayumn")

                ress, err := http.DefaultClient.Do(reqq)
                if err != nil {
                        panic(err)
                }

                defer ress.Body.Close() 

                data, err := ioutil.ReadAll(ress.Body)
                if err != nil {
                        fmt.Println("error! :", err)
                }

                fmt.Println("Data: ", string(data)) */

                req, err := http.NewRequest("POST", "https://api.logging.eu-west-1.prod.firetail.app/logs/bulk", bytes.NewBuffer(reqBytes))
		if err != nil {
			return err
		}

		req.Header.Set("x-ft-api-key", options.LogApiKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
		//	return err
                        panic(err)
		}

		defer resp.Body.Close()

		respStr, err := ioutil.ReadAll(resp.Body)
		if err != nil {
                        return err
		}
                log.Println("respBODY: ", string(respStr))

		/*var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		if res["message"] != "success" {
			return errors.New(fmt.Sprintf("got err response from firetail api: %v", res))
		}*/ 

		return nil
	}

	return func(batch [][]byte) {
		// If there's no log API url or log API key set then we can't log, so just return
		if options.LogApiUrl == "" || options.LogApiKey == "" {
			return
		}

		var err error
		retries := 0
		for {
			err = sendBatch(batch)
			retries++
			// If sendBatch succeeded, or we've had 3 retries, we give up
			if err == nil || retries >= 3 {
				break
			}
		}
	}
}
