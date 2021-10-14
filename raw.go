package bamboo

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// RawService handles communication with the raw request service
type RawService service

// RawResponse encapsultes a raw response
type RawResponse struct {
	Body string
}

// GetRaw will send a get request
func (p *RawService) GetRaw(path string) (string, *http.Response, error) {

	path = strings.TrimPrefix(path, p.client.BaseUrl.String())

	request, err := p.client.RawRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", nil, err
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := p.client.RawDo(request, nil)
	if err != nil {
		return "", nil, err
	}

	if !(response.StatusCode == 200) {
		return "", nil, &simpleError{fmt.Sprintf("Get returned %d", response.StatusCode)}
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", response, &simpleError{fmt.Sprintf("Read body %s", err)}
	}

	return string(body), response, nil
}
