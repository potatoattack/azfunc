package triggers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/KarlGW/azfunc/data"
)

var (
	// ErrHTTPInvalidContentType is returned when an invalid Content-Type provided.
	ErrHTTPInvalidContentType = errors.New("invalid Content-Type")
	// ErrHTTPInvalidBody is returned when the HTTP body is invalid.
	ErrHTTPInvalidBody = errors.New("invalid body")
)

// HTTP represents an HTTP trigger.
type HTTP struct {
	URL        string
	Method     string
	Body       data.Raw
	Headers    http.Header
	Params     map[string]string
	Query      map[string]string
	Identities []HTTPIdentity
	Metadata   HTTPMetadata
}

// HTTPMetadata represents the metadata for an HTTP trigger.
type HTTPMetadata struct {
	Headers map[string]string
	Params  map[string]string
	Query   map[string]string
	Metadata
}

// HTTPIdentity represent a part of the Identities field
// of the incoming trigger request.
type HTTPIdentity struct {
	IsAuthenticated    bool
	AuthenticationType string
	NameClaimType      string
	RoleClaimType      string
	Actor              any
	BootstrapContext   any
	Label              any
	Name               any
	Claims             []HTTPIdentityClaims
}

// HTTPIdentityClaims represent the claims of an HTTPIdentity.
type HTTPIdentityClaims struct {
	Issuer         string
	OriginalIssuer string
	Type           string
	Value          string
	ValueType      string
	Properties     map[string]string
}

// Parse the body from the HTTP trigger into the provided value.
func (t HTTP) Parse(v any) error {
	return json.Unmarshal(t.Body, &v)
}

// Data returns the Raw data of the HTTP trigger.
func (t HTTP) Data() data.Raw {
	return t.Body
}

// FormData parses the HTTP trigger for form data sent with Content-Type
// application/x-www-form-urlencoded and returns it as url.Values.
func (t HTTP) FormData() (url.Values, error) {
	contentType := t.Headers.Get("Content-Type")
	if strings.ToLower(contentType) != "application/x-www-form-urlencoded" {
		return nil, fmt.Errorf("%w: %s", ErrHTTPInvalidContentType, contentType)
	}

	data, err := url.ParseQuery(string(t.Body))
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrHTTPInvalidBody, string(t.Body))
	}
	if len(data) == 1 {
		for _, v := range data {
			if len(v[0]) == 0 {
				return nil, fmt.Errorf("%w: %s", ErrHTTPInvalidBody, string(t.Body))
			}
		}
	}

	return data, nil
}

// NewHTTP creates and returns an HTTP trigger from the provided
// *http.Request.
func NewHTTP(r *http.Request, options ...Option) (*HTTP, error) {
	opts := Options{}
	for _, option := range options {
		option(&opts)
	}

	var t httpTrigger
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return nil, ErrTriggerPayloadMalformed
	}
	defer r.Body.Close()

	return &HTTP{
		URL:        t.Data.Req.URL,
		Method:     t.Data.Req.Method,
		Body:       t.Data.Req.Body,
		Headers:    t.Data.Req.Headers,
		Params:     t.Data.Req.Params,
		Query:      t.Data.Req.Query,
		Identities: t.Data.Req.Identities,
		Metadata:   t.Metadata,
	}, nil
}

// httpTrigger is the incoming request from the Function host.
type httpTrigger struct {
	Data struct {
		Req struct {
			URL        string `json:"Url"`
			Method     string
			Body       data.Raw
			Headers    http.Header
			Params     map[string]string
			Query      map[string]string
			Identities []HTTPIdentity
		} `json:"req"`
	}
	Metadata HTTPMetadata
}
