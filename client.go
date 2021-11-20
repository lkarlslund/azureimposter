package azureimposter

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

type Client struct {
	token Token
	*resty.Client
	OnTokenRefresh func(Token)
}

type AzureRequest struct {
	client *Client
	*resty.Request
}

func NewClient(token Token) *Client {
	client := &Client{
		token:  token,
		Client: resty.New(),
	}

	client.RetryCount = 10
	client.AddRetryCondition(func(resp *resty.Response, err error) bool {
		if resp.StatusCode() == 429 && resp.Request.Attempt < 10 {
			// Throttling
			return true
		}
		if gjson.GetBytes(resp.Body(), "error").Exists() && gjson.GetBytes(resp.Body(), "error.code").String() == "InvalidAuthenticationToken" {
			// Token expired, refresh
			err := client.token.Refresh()
			if err == nil {
				client.SetAuthToken(client.token.AccessToken)       // For the next requests
				resp.Request.SetAuthToken(client.token.AccessToken) // For this request
				if client.OnTokenRefresh != nil {
					client.OnTokenRefresh(client.token)
				}
				return true
			}
			return false // Token refresh failed
		}
		return false
	})
	client.RetryAfter = func(c *resty.Client, r *resty.Response) (time.Duration, error) {
		retryafter := r.Header().Get("Retry-After")
		if retryafter != "" {
			delay, err := strconv.ParseFloat(r.Header().Get("Retry-After"), 64)
			if err == nil {
				return time.Duration(delay*1000) * time.Millisecond, nil
			}
		}
		return 0, nil
	}
	client.SetAuthToken(client.token.AccessToken)
	return client
}

func (a *Client) R() *AzureRequest {
	return &AzureRequest{
		Request: a.Client.R(),
		client:  a,
	}
}
func (ar *AzureRequest) autoRefreshToken() error {
	if !ar.client.token.IsValid() {
		return ar.client.token.Refresh()
	}
	return nil
}

func (ar *AzureRequest) GetData(onData func(data []byte) error) error {
	ar.autoRefreshToken()

	res, err := ar.Send()
	if err != nil {
		// Handle token refresh here
		return err
	}

	// Was there an error?
	if gjson.GetBytes(res.Body(), "error").Exists() {
		return fmt.Errorf("Problem getting %v: %w", ar.URL, errors.New(gjson.GetBytes(res.Body(), "error.code").String()))
	}

	body := res.Body()

	err = onData([]byte(body))
	return err
}

func (ar *AzureRequest) GetChunkedData(onChunk func(data []byte) error) error {
	ar.autoRefreshToken()

	res, err := ar.Send()
	if err != nil {
		// Handle token refresh here
		return err
	}

	for {
		// Was there an error?
		if gjson.GetBytes(res.Body(), "error").Exists() {
			return fmt.Errorf("Problem getting %v: %w", ar.URL, errors.New(gjson.GetBytes(res.Body(), "error.code").String()))
		}

		body := res.Body()

		// debug
		// fmt.Println(gjson.GetBytes(body, "value").Raw)
		value := gjson.GetBytes(body, "value")
		if value.Exists() {
			err = onChunk([]byte(value.Raw))
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("No value in Azure response: %v", string(body))
		}
		if nextlink := gjson.GetBytes(body, "@odata\\.nextLink").String(); nextlink != "" {
			req := ar.client.R()
			res, err = req.Get(nextlink)
		} else {
			break
		}
	}

	return nil
}

func (ar *AzureRequest) BatchChunkData(requests []BatchRequest, onRequest func(data []byte) error) error {
	ar.autoRefreshToken()

	ar.Header.Add("Accept", "application/json")
	ar.Header.Add("Content-Type", "application/json")
	type wrapper struct {
		Requests []BatchRequest `json:"requests"`
	}
	ar.SetBody(wrapper{Requests: requests})

	res, err := ar.Send()
	if err != nil {
		// Handle token refresh here
		return err
	}

	type unwrapper struct {
		Responses []BatchResponse `json:"responses"`
	}

	for {
		// Was there an error?
		if gjson.GetBytes(res.Body(), "error").Exists() {
			return fmt.Errorf("Problem getting %v: %w", ar.URL, errors.New(gjson.GetBytes(res.Body(), "error.code").String()))
		}

		body := res.Body()

		err = onRequest([]byte(gjson.GetBytes(body, "value").Raw))
		if err != nil {
			return err
		}

		if nextlink := gjson.GetBytes(body, "@odata\\.nextLink").String(); nextlink != "" {
			req := ar.client.R()
			res, err = req.Get(nextlink)
		} else {
			break
		}
	}

	return nil
}

type BatchRequest struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Method string `json:"method"`
	Body   string `json:"body"`
}

type BatchResponse struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Method string `json:"method"`
	Body   string `json:"body"`
}
