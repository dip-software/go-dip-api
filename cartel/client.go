// Package cartel provides support for HSDP Cartel services
package cartel

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-querystring/query"

	"github.com/philips-software/go-hsdp-api/fhir"
)

const (
	libraryVersion = "0.15.0"
	userAgent      = "go-hsdp-api/cartel/" + libraryVersion
)

var (
	ErrMissingSecret  = errors.New("missing secret")
	ErrMissingToken   = errors.New("missing token")
	ErrMissingHost    = errors.New("missing host")
	ErrNotImplemented = errors.New("not implemented")
)

// Config the client
type Config struct {
	SkipVerify bool
	Token      string
	Secret     []byte
	NoTLS      bool
	Host       string
	Debug      bool
}

// Valid returns if all required config fields are present, false otherwise
func (c *Config) Valid() (bool, error) {
	if len(c.Secret) == 0 {
		return false, ErrMissingSecret
	}
	if len(c.Token) == 0 {
		return false, ErrMissingToken
	}
	if c.Host == "" {
		return false, ErrMissingHost
	}
	return true, nil
}

// Client holds the client state
type Client struct {
	config     Config
	httpClient *http.Client
	baseURL    *url.URL
	userAgent  string
}

// Response holds a LogEvent response
type Response struct {
	*http.Response
	Message string
}

// OptionFunc is the function signature function for options
type OptionFunc func(*http.Request) error

// newResponse creates a new Response for the provided http.Response.
func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	return response
}

// NewClient returns an instance of the logger client with the given Config
func NewClient(httpClient *http.Client, config Config) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if valid, err := config.Valid(); !valid {
		return nil, err
	}
	var cartel Client

	cartel.config = config
	cartel.httpClient = httpClient
	cartel.userAgent = userAgent

	// Make sure the given URL end with a slash
	if !strings.HasSuffix(cartel.config.Host, "/") {
		cartel.config.Host += "/"
	}
	var err error
	cartel.baseURL, err = url.Parse(cartel.config.Host)
	if err != nil {
		return nil, err
	}

	if os.Getenv("DEBUG") == "true" {
		cartel.config.Debug = true
	}
	return &cartel, nil
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
func (c *Client) Do(req *http.Request, v interface{}) (*Response, error) {
	if c.config.Debug {
		dumped, _ := httputil.DumpRequest(req, true)
		fmt.Fprintf(os.Stderr, "REQUEST: %s\n", string(dumped))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := newResponse(resp)

	if c.config.Debug {
		if resp != nil {
			dumped, _ := httputil.DumpResponse(resp, true)
			fmt.Fprintf(os.Stderr, "RESPONSE: %s\n", string(dumped))
		} else {
			fmt.Fprintf(os.Stderr, "Error sending response: %s\n", err)
		}
	}

	err = CheckResponse(resp)
	if err != nil {
		// even though there was an error, we still return the response
		// in case the caller wants to inspect it further
		return response, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
		}
	}

	return response, err
}

// ErrorResponse holds an error response from the server
type ErrorResponse struct {
	Response *http.Response
	Message  string
}

func (e *ErrorResponse) Error() string {
	path, _ := url.QueryUnescape(e.Response.Request.URL.Opaque)
	u := fmt.Sprintf("%s://%s%s", e.Response.Request.URL.Scheme, e.Response.Request.URL.Host, path)
	return fmt.Sprintf("%s %s: %d %s", e.Response.Request.Method, u, e.Response.StatusCode, e.Message)
}

// CheckResponse checks the API response for errors, and returns them if present.
func CheckResponse(r *http.Response) error {
	switch r.StatusCode {
	case 200, 201, 202, 204, 304:
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		var raw interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			errorResponse.Message = "failed to parse unknown error format"
		}

		errorResponse.Message = fhir.ParseError(raw)
	}

	return errorResponse
}

// NewRequest creates an API request. A relative URL path can be provided in
// urlStr, in which case it is resolved relative to the base URL of the Client.
// Relative URL paths should always be specified without a preceding slash. If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (c *Client) NewRequest(method, path string, opt interface{}, options []OptionFunc) (*http.Request, error) {
	var u url.URL

	u = *c.baseURL
	u.Opaque = c.baseURL.Path + path

	if opt != nil {
		q, err := query.Values(opt)
		if err != nil {
			return nil, err
		}
		u.RawQuery = q.Encode()
	}

	req := &http.Request{
		Method:     method,
		URL:        &u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       u.Host,
	}

	for _, fn := range options {
		if fn == nil {
			continue
		}

		if err := fn(req); err != nil {
			return nil, err
		}
	}

	if method == "POST" || method == "PUT" {
		bodyBytes, err := json.Marshal(opt)
		if err != nil {
			return nil, err
		}
		bodyReader := bytes.NewReader(bodyBytes)

		u.RawQuery = ""
		req.Body = ioutil.NopCloser(bodyReader)
		req.ContentLength = int64(bodyReader.Len())
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", string(generateSignature(c.config.Secret, bodyBytes)))
	}
	req.Header.Set("Accept", "application/json")

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	return req, nil
}

func generateSignature(secret []byte, payload []byte) string {
	hash := hmac.New(sha256.New, secret)
	hash.Write(payload)
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}
