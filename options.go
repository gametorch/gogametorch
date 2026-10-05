package gametorch

import (
	"net/http"
	"time"
)

type config struct {
	baseURL        string
	apiKey         string
	bearerToken    string
	timeout        time.Duration
	connectTimeout time.Duration
	maxRetries     int
	retryBaseDelay time.Duration
	rateLimit      bool
	userAgent      string
	defaultHeaders http.Header
	httpClient     *http.Client
}

func defaultConfig() config {
	return config{
		timeout:        120 * time.Second,
		connectTimeout: 15 * time.Second,
		maxRetries:     3,
		retryBaseDelay: 500 * time.Millisecond,
		rateLimit:      true,
		userAgent:      "gametorch-go/" + Version,
		defaultHeaders: make(http.Header),
	}
}

// Option configures a Client.
type Option func(*config) error

// WithAPIKey sets a gt2_... API key used for server-to-server authentication.
func WithAPIKey(apiKey string) Option {
	return func(c *config) error {
		c.apiKey = apiKey
		c.bearerToken = ""
		return nil
	}
}

// WithBearerToken sets a Clerk session/bearer token.
func WithBearerToken(token string) Option {
	return func(c *config) error {
		c.bearerToken = token
		c.apiKey = ""
		return nil
	}
}

// WithBaseURL overrides the base URL. It defaults to DefaultBaseURL. For local
// development use "http://localhost:8300/api".
func WithBaseURL(baseURL string) Option {
	return func(c *config) error {
		c.baseURL = baseURL
		return nil
	}
}

// WithTimeout sets the total request timeout (default 120s). A zero duration
// disables the timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *config) error {
		c.timeout = timeout
		return nil
	}
}

// WithNoTimeout disables the total request timeout.
func WithNoTimeout() Option {
	return func(c *config) error {
		c.timeout = 0
		return nil
	}
}

// WithConnectTimeout sets the connection timeout (default 15s).
func WithConnectTimeout(timeout time.Duration) Option {
	return func(c *config) error {
		c.connectTimeout = timeout
		return nil
	}
}

// WithMaxRetries sets the maximum number of automatic retries for retryable
// failures (default 3). Set to 0 to disable retries.
func WithMaxRetries(maxRetries int) Option {
	return func(c *config) error {
		c.maxRetries = maxRetries
		return nil
	}
}

// WithRetryBaseDelay sets the base delay for exponential backoff between
// retries (default 500ms).
func WithRetryBaseDelay(delay time.Duration) Option {
	return func(c *config) error {
		c.retryBaseDelay = delay
		return nil
	}
}

// WithRateLimit enables or disables the built-in client-side rate limiting and
// concurrency caps (enabled by default).
func WithRateLimit(enabled bool) Option {
	return func(c *config) error {
		c.rateLimit = enabled
		return nil
	}
}

// WithUserAgent overrides the User-Agent header.
func WithUserAgent(userAgent string) Option {
	return func(c *config) error {
		c.userAgent = userAgent
		return nil
	}
}

// WithDefaultHeader adds a default header sent with every request.
func WithDefaultHeader(name, value string) Option {
	return func(c *config) error {
		c.defaultHeaders.Set(name, value)
		return nil
	}
}

// WithHTTPClient supplies a custom *http.Client. When set, the timeout and
// connect-timeout options are ignored.
func WithHTTPClient(client *http.Client) Option {
	return func(c *config) error {
		c.httpClient = client
		return nil
	}
}
