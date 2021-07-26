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
	client.SetDebug(config.Debug)
	if MockHttpClientFunc != nil {
		MockHttpClientFunc(client.GetClient())
	}
	return client, nil
}

func NewApiRequest(config Config, state State) (*resty.Request, error) {
	err := state.RequireAuthToken()
	if err != nil {
		return nil, err
	}

	r, err := NewApiRequestWithoutAuth(config)
	if err != nil {
		return nil, err
	}

	r.SetAuthToken(state.AuthToken)
	return r, nil
}

func NewApiRequestWithoutAuth(config Config) (*resty.Request, error) {
	client, err := NewApiClient(config)
	if err != nil {
		return nil, err
	}

	r := client.R()
	r.SetError(&map[string]interface{}{})
	r.SetHeader("User-Agent", "sqedule-cli")
	r.SetHeader("Accept", "application/json")
	return r, nil
}

func GetApiErrorMessage(resp *resty.Response) string {
	e := resp.Error()
	if e == nil {
		return fmt.Sprintf("unknown error (HTTP %s)", resp.Status())
	}

	object, ok := e.(*map[string]interface{})
	if !ok {
		return fmt.Sprintf("HTTP %s: %#v", resp.Status(), object)
	}
	if object == nil || *object == nil {
		return fmt.Sprintf("unknown error (HTTP %s)", resp.Status())
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

	if len(*object) == 0 {
		return fmt.Sprintf("unknown error (HTTP %s)", resp.Status())
	} else {
		return fmt.Sprintf("HTTP %s: %#v", resp.Status(), object)
	}
}
