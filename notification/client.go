// Package notification provides support for interacting with HSDP Notification services
package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	autoconf "github.com/dip-software/go-dip-api/config"
	"github.com/dip-software/go-dip-api/iam"
	"github.com/dip-software/go-dip-api/internal"
	"github.com/go-playground/validator/v10"
	"github.com/google/go-querystring/query"
)

const (
	userAgent  = "go-dip-api/notification/" + internal.LibraryVersion
	APIVersion = "2"
)

// OptionFunc is the function signature function for options
type OptionFunc func(*http.Request) error

// Config contains the configuration of a client
type Config struct {
	Region          string
	Environment     string
	OrganizationID  string
	NotificationURL string
	Type            string
	TimeZone        string
	DebugLog        io.Writer
	Retry           int
}

// A Client manages communication with HSDP Notification API
type Client struct {
	// HTTP client used to communicate with IAM API
	iamClient *iam.Client

	config *Config

	notificationURL *url.URL

	// User agent used when communicating with the HSDP Notification API
	UserAgent string

	validate *validator.Validate

	Producer     *ProducerService
	Subscription *SubscriptionService
	Subscriber   *SubscriberService
	Topic        *TopicService
}

// NewClient returns a new HSDP Notification API client. A configured IAM client
// must be provided as the underlying API requires an IAM token
func NewClient(iamClient *iam.Client, config *Config) (*Client, error) {
	return newClient(iamClient, config)
}

func newClient(iamClient *iam.Client, config *Config) (*Client, error) {
	doAutoconf(config)
	c := &Client{iamClient: iamClient, config: config, UserAgent: userAgent, validate: validator.New()}

	if err := c.SetNotificationURL(config.NotificationURL); err != nil {
		return nil, err
	}

	c.Producer = &ProducerService{client: c, validate: validator.New()}
	c.Subscriber = &SubscriberService{client: c, validate: validator.New()}
	c.Subscription = &SubscriptionService{client: c, validate: validator.New()}
	c.Topic = &TopicService{client: c, validate: validator.New()}

	return c, nil
}

func doAutoconf(config *Config) {
	if config.Region != "" && config.Environment != "" {
		c, err := autoconf.New(
			autoconf.WithRegion(config.Region),
			autoconf.WithEnv(config.Environment))
		if err == nil {
			notificationService := c.Service("notification")
			if notificationService.URL != "" && config.NotificationURL == "" {
				config.NotificationURL = notificationService.URL
			}
		}
	}
}

// Close releases allocated resources of clients
func (c *Client) Close() {
}

// SetNotificationURL sets the Notification URL for API requests
func (c *Client) SetNotificationURL(urlStr string) error {
	if urlStr == "" {
		return ErrNotificationURLCannotBeEmpty
	}
	// Make sure the given URL end with a slash
	if !strings.HasSuffix(urlStr, "/") {
		urlStr += "/"
	}
	var err error
	c.notificationURL, err = url.Parse(urlStr)
	return err
}

// newNotificationRequest creates an new Notification Service API request. A relative URL path can be provided in
// urlStr, in which case it is resolved relative to the base URL of the Client.
// Relative URL paths should always be specified without a preceding slash. If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (c *Client) newNotificationRequest(method, path string, opt interface{}, options ...OptionFunc) (*http.Request, error) {
	u := *c.notificationURL
	// Set the encoded opaque data
	u.Opaque = c.notificationURL.Path + path

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
	if opt != nil {
		q, err := query.Values(opt)
		if err != nil {
			return nil, err
		}
		u.RawQuery = q.Encode()
	}

	if method == "POST" || method == "PUT" {
		bodyBytes, err := json.Marshal(opt)
		if err != nil {
			return nil, err
		}
		bodyReader := bytes.NewReader(bodyBytes)

		u.RawQuery = ""
		req.Body = io.NopCloser(bodyReader)
		req.ContentLength = int64(bodyReader.Len())
		req.Header.Set("Content-Type", "application/json")
	}
	token, err := c.iamClient.Token()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("API-Version", APIVersion)
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	for _, fn := range options {
		if fn == nil {
			continue
		}

		if err := fn(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// Response is a HSDP IAM API response. This wraps the standard http.Response
// returned from HSDP IAM and provides convenient access to things like errors
type Response struct {
	*http.Response
}

// newResponse creates a new Response for the provided http.Response.
func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	return response
}

// TokenRefresh forces a refresh of the IAM access token
func (c *Client) TokenRefresh() error {
	if c.iamClient == nil {
		return fmt.Errorf("invalid IAM client, cannot refresh token")
	}
	return c.iamClient.TokenRefresh()
}

// do executes a http request. If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
func (c *Client) do(req *http.Request, v interface{}) (*Response, error) {
	resp, err := c.iamClient.HttpClient().Do(req)
	if err != nil {
		return nil, err
	}

	response := newResponse(resp)

	err = internal.CheckResponse(resp)
	if err != nil {
		// even though there was an error, we still return the response
		// in case the caller wants to inspect it further
		return response, err
	}

	if v != nil {
		defer resp.Body.Close() // Only close if we plan to read it
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
		}
	}
	return response, err
}
