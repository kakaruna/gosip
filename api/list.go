package api

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/koltyakov/gosip"
)

// List ...
type List struct {
	client    *gosip.SPClient
	config    *RequestConfig
	endpoint  string
	modifiers map[string]string
}

// Conf ...
func (list *List) Conf(config *RequestConfig) *List {
	list.config = config
	return list
}

// Select ...
func (list *List) Select(oDataSelect string) *List {
	if list.modifiers == nil {
		list.modifiers = make(map[string]string)
	}
	list.modifiers["$select"] = oDataSelect
	return list
}

// Expand ...
func (list *List) Expand(oDataExpand string) *List {
	if list.modifiers == nil {
		list.modifiers = make(map[string]string)
	}
	list.modifiers["$expand"] = oDataExpand
	return list
}

// Get ...
func (list *List) Get() ([]byte, error) {
	apiURL, _ := url.Parse(list.endpoint)
	query := url.Values{}
	for k, v := range list.modifiers {
		query.Add(k, trimMultiline(v))
	}
	apiURL.RawQuery = query.Encode()
	sp := NewHTTPClient(list.client)
	return sp.Get(apiURL.String(), getConfHeaders(list.config))
}

// Delete ...
func (list *List) Delete() ([]byte, error) {
	sp := NewHTTPClient(list.client)
	return sp.Delete(list.endpoint, getConfHeaders(list.config))
}

// Update ...
func (list *List) Update(body []byte) ([]byte, error) {
	sp := NewHTTPClient(list.client)
	return sp.Update(list.endpoint, body, getConfHeaders(list.config))
}

// Items ...
func (list *List) Items() *Items {
	return &Items{
		client:   list.client,
		config:   list.config,
		endpoint: fmt.Sprintf("%s/items", list.endpoint),
	}
}

// GetEntityType ...
func (list *List) GetEntityType() (string, error) {
	headers := getConfHeaders(list.config)
	headers["Accept"] = "application/json;odata=verbose"

	data, err := list.Select("ListItemEntityTypeFullName").Conf(&RequestConfig{Headers: headers}).Get()
	if err != nil {
		return "", err
	}

	res := &struct {
		D struct {
			Results struct {
				ListItemEntityTypeFullName string `json:"ListItemEntityTypeFullName"`
			} `json:"results"`
		} `json:"d"`
	}{}

	if err := json.Unmarshal(data, &res); err != nil {
		return "", fmt.Errorf("unable to parse the response: %v", err)
	}

	return res.D.Results.ListItemEntityTypeFullName, nil
}

// Roles ...
func (list *List) Roles() *Roles {
	return &Roles{
		client:   list.client,
		config:   list.config,
		endpoint: list.endpoint,
	}
}
