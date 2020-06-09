package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func errorFrom(res *http.Response) error {
	defer res.Body.Close()

	var out struct {
		Error string `json:"error"`
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	return fmt.Errorf(out.Error)
}
