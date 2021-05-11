package cli

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

var MockHttpClientFunc func(client *http.Client)

func NewApiClient(config Config) (*resty.Client, error) {
	err := config.RequireServerConfig()
	if err != nil {
		return nil, err
	}

	client := resty.New()
	client.SetHostURL(config.ServerBaseURL + "/v1")
	if MockHttpClientFunc != nil {
		MockHttpClientFunc(client.GetClient())
	}
	return client, nil
}

func NewApiRequest(config Config, state State) (*resty.Request, error) {
	client, err := NewApiClient(config)
	if err != nil {
		return nil, err
	}

	r := client.R()
	r.SetError(&map[string]interface{}{})
	r.SetAuthToken(state.AuthToken)
	r.SetHeader("User-Agent", "sqedule-cli")
	r.SetHeader("Accept", "application/json")
	return r, nil
}

func GetApiErrorMessage(resp *resty.Response) string {
	e := resp.Error()
	if e == nil {
		return "unknown error"
	}

	object, ok := e.(*map[string]interface{})
	if !ok {
		return fmt.Sprintf("%#v", object)
	}
	if object == nil {
		return "unknown error"
	}

	if result, ok := (*object)["message"]; ok {
		if str, ok := result.(string); ok {
			return str
		} else {
			return fmt.Sprintf("%#v", result)
		}
	}

	if result, ok := (*object)["error"]; ok {
		if str, ok := result.(string); ok {
			return str
		} else {
			return fmt.Sprintf("%#v", result)
		}
	}

	return fmt.Sprintf("%#v", object)
}
