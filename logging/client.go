// Package logging provides support for HSDP Logging services
package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/dip-software/go-dip-api/internal"

	"github.com/dip-software/go-dip-api/iam"

	autoconf "github.com/dip-software/go-dip-api/config"

	signer "github.com/dip-software/go-dip-signer"
)

const (
	// TimeFormat is the time format used for the LogTime field
	TimeFormat = "2006-01-02T15:04:05.000Z07:00"

	userAgent = "go-dip-api/logging/" + internal.LibraryVersion
)

var (
	entryRegex = regexp.MustCompile(`^entry\[(\d+)]`)

	scaryMap = map[string]string{
		";":    "[sc]",
		"&":    "[amp]",
		">":    "[gt]",
		"<":    "[lt]",
		"\\r":  "",
		"\\u":  "[utf]",
		"\\f":  "[ff]",
		"\\\"": "[qt]",
	}
)

type replacer struct {
	Regexp  *regexp.Regexp
	Replace map[string]string
}

func (r replacer) replace(input string) string {
	for k, v := range r.Replace {
		input = strings.ReplaceAll(input, k, v)
	}
	return input
}

var (
	replacerMap = map[string]replacer{
		"applicationVersion": {
			Regexp: regexp.MustCompile(`^[^&+;=?@|<>()]*$`),
			Replace: map[string]string{
				"@": "💀",
				"&": "💀",
				"+": "💀",
				";": "💀",
				"=": "💀",
				"?": "💀",
				"<": "💀",
				">": "💀",
				"|": "💀",
				"(": "💀",
				")": "💀",
			},
		},
	}
)

// Config the client
type Config struct {
	Region       string
	Environment  string
	SharedKey    string
	SharedSecret string
	IAMClient    *iam.Client
	BaseURL      string
	ProductKey   string
	Debug        bool
	DebugLog     io.Writer
}

// Valid returns if all required config fields are present, false otherwise
func (c *Config) Valid() (bool, error) {
	if c.SharedKey == "" && c.IAMClient == nil {
		return false, ErrMissingSharedKey
	}
	if c.SharedSecret == "" && c.IAMClient == nil {
		return false, ErrMissingSharedSecret
	}
	if c.BaseURL == "" {
		return false, ErrMissingBaseURL
	}
	if c.ProductKey == "" {
		return false, ErrMissingProductKey
	}
	return true, nil
}

// Client holds the client state
type Client struct {
	*iam.Client
	config     *Config
	url        *url.URL
	httpClient *http.Client
	httpSigner *signer.Signer
}

// StoreResponse holds a LogEvent response
type StoreResponse struct {
	*http.Response
	Message string
	Failed  []Resource
}

// CustomIndexBody describes the custom index request payload
type CustomIndexBody []struct {
	Fieldname string `json:"fieldname"`
	Fieldtype string `json:"fieldtype"`
}

func (r *StoreResponse) StatusCode() int {
	if r.Response != nil {
		return r.Response.StatusCode
	}
	return 0
}

// NewClient returns an instance of the logger client with the given Config
func NewClient(httpClient *http.Client, config *Config) (*Client, error) {
	var logger Client

	if httpClient == nil {
		c := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		}
		httpClient = c
	}
	if config.DebugLog != nil {
		httpClient.Transport = internal.NewLoggingRoundTripper(httpClient.Transport, config.DebugLog)
	}
	// Autoconfig
	if config.Region != "" && config.Environment != "" {
		c, err := autoconf.New(
			autoconf.WithRegion(config.Region),
			autoconf.WithEnv(config.Environment))
		if err == nil {
			loggingService := c.Service("logging")
			if loggingService.URL != "" && config.BaseURL == "" {
				config.BaseURL = loggingService.URL
			}
		}
	}
	if valid, err := config.Valid(); !valid {
		return nil, err
	}

	logger.config = config
	logger.httpClient = httpClient

	parsedURL, err := url.Parse(config.BaseURL + "/core/log/LogEvent")
	if err != nil {
		return nil, err
	}

	logger.httpSigner, err = signer.New(logger.config.SharedKey, logger.config.SharedSecret)
	if err != nil {
		if config.IAMClient == nil {
			return nil, ErrMissingCredentialsOrIAMClient
		}
		logger.Client = config.IAMClient
	}

	logger.url = parsedURL
	if os.Getenv("DEBUG") == "true" {
		logger.config.Debug = true
	}
	return &logger, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}
	if err != nil {
		return nil, err
	}
	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
		}
	} else {
		_, _ = io.ReadAll(resp.Body)
	}
	return resp, err
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

// StoreResources posts one or more log messages
// In case invalid resources are detected StoreResources will return
// with ErrBatchErrors and the Response.Failed map will contain the resources
// This also happens in case the HSDP Ingestor API flags resources. In both cases
// the complete batch should be considered as not persisted and the LogEvents should
// be resubmitted for storage
func (c *Client) StoreResources(msgs []Resource, count int) (*StoreResponse, error) {
	b := Bundle{
		ResourceType: "Bundle",
		Type:         "transaction",
		ProductKey:   c.config.ProductKey,
	}
	var invalid []Resource

	j := 0
	for i := 0; i < count; i++ {
		msg := msgs[i]
		replaceScaryCharacters(&msg)
		if !msg.Valid() {
			invalid = append(invalid, msg)
			continue
		}
		// Element
		var e Element
		e.Resource = msg
		e.Resource.ResourceType = "LogEvent"
		b.Entry = append(b.Entry, e)
		j++
	}
	if len(invalid) > 0 { // Don't even POST anything due to errors in the batch
		resp := StoreResponse{
			Failed: invalid,
		}
		return &resp, ErrBatchErrors
	}
	b.Total = j

	bodyBytes, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(bodyBytes)

	req, err := http.NewRequest(http.MethodPost, c.url.String(), bodyReader)
	if err != nil {
		return nil, err
	}
	req.ContentLength = int64(bodyReader.Len())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Version", "1")
	req.Header.Set("User-Agent", userAgent)
	if c.httpSigner != nil {
		if err := c.httpSigner.SignRequest(req); err != nil {
			return nil, err
		}
	} else {
		token, err := c.Token()
		if err != nil {
			req.Header.Set("X-Token-Error", fmt.Sprintf("%v", err))
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	var serverResponse bytes.Buffer

	resp, err := c.do(req, &serverResponse)
	if resp == nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	defer func() {
		_ = req.Body.Close()
	}()
	storeResp := &StoreResponse{Response: resp}
	if resp.StatusCode != http.StatusCreated { // Only good outcome
		var errResponse bundleErrorResponse
		err := json.Unmarshal(serverResponse.Bytes(), &errResponse)
		if err != nil {
			return storeResp, err
		}
		if len(errResponse.Issue) == 0 || len(errResponse.Issue[0].Location) == 0 {
			return storeResp, ErrResponseError
		}
		for _, entry := range errResponse.Issue[0].Location {
			if entries := entryRegex.FindStringSubmatch(entry); len(entries) > 1 {
				i, err := strconv.Atoi(entries[1])
				if err != nil {
					return storeResp, err
				}
				invalidResource := msgs[i]
				invalidResource.Error = fmt.Errorf("issue location %s", entry)
				invalid = append(invalid, invalidResource)
			}
		}
	}
	if len(invalid) > 0 {
		storeResp.Failed = invalid
		err = ErrBatchErrors
	}
	return storeResp, err
}

func replaceScaryCharacters(msg *Resource) {
	// Application version fixer
	appVersion := replacerMap["applicationVersion"]
	if !appVersion.Regexp.MatchString(msg.ApplicationVersion) {
		msg.ApplicationVersion = appVersion.replace(msg.ApplicationVersion)
	}

	if len(msg.Custom) == 0 {
		return
	}
	stringCustom := strings.Replace(string(msg.Custom), "\\\\", "[bsl]", -1)

	for s, r := range scaryMap {
		stringCustom = strings.Replace(stringCustom, s, r, -1)
	}
	msg.Custom = []byte(stringCustom)
}
