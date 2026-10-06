package gametorch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Version is the SDK version.
const Version = "0.1.1"

// DefaultBaseURL is the production GameTorch API base URL.
const DefaultBaseURL = "https://gametorch.app/api"

// EnvAPIKey is the environment variable read by NewClientFromEnv for the API
// key or bearer token.
const EnvAPIKey = "GAMETORCH_API_KEY"

// EnvBaseURL is the environment variable read by NewClientFromEnv for the base
// URL.
const EnvBaseURL = "GAMETORCH_BASE_URL"

// Client is an HTTP client for the GameTorch API. It is safe for concurrent
// use and cheap to copy; all copies share the underlying connection pool and
// rate limiter.
type Client struct {
	http           *http.Client
	baseURL        *url.URL
	authToken      string
	limiter        *rateLimiter
	maxRetries     int
	retryBaseDelay time.Duration
	userAgent      string
	defaultHeaders http.Header
}

// NewClient builds a client from the supplied options.
func NewClient(opts ...Option) (*Client, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return nil, asError(err)
		}
	}
	return newClient(cfg)
}

// NewClientFromEnv builds a client from the GAMETORCH_API_KEY and optional
// GAMETORCH_BASE_URL environment variables, then applies any options, which
// take precedence.
func NewClientFromEnv(opts ...Option) (*Client, error) {
	var envOpts []Option
	if apiKey := os.Getenv(EnvAPIKey); apiKey != "" {
		envOpts = append(envOpts, WithAPIKey(apiKey))
	}
	if baseURL := os.Getenv(EnvBaseURL); baseURL != "" {
		envOpts = append(envOpts, WithBaseURL(baseURL))
	}
	envOpts = append(envOpts, opts...)
	return NewClient(envOpts...)
}

func newClient(cfg config) (*Client, error) {
	base := cfg.baseURL
	if base == "" {
		base = DefaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return nil, newInvalidBaseURLError(err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, newInvalidBaseURLError(fmt.Errorf("missing scheme or host in %q", base))
	}
	if !strings.HasSuffix(parsed.Path, "/") {
		parsed.Path = strings.TrimSuffix(parsed.Path, "/") + "/"
	}

	httpClient := cfg.httpClient
	if httpClient == nil {
		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   cfg.connectTimeout,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   16,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		httpClient = &http.Client{Transport: transport, Timeout: cfg.timeout}
	}

	token := cfg.apiKey
	if token == "" {
		token = cfg.bearerToken
	}

	return &Client{
		http:           httpClient,
		baseURL:        parsed,
		authToken:      token,
		limiter:        newRateLimiter(cfg.rateLimit),
		maxRetries:     cfg.maxRetries,
		retryBaseDelay: cfg.retryBaseDelay,
		userAgent:      cfg.userAgent,
		defaultHeaders: cfg.defaultHeaders,
	}, nil
}

// BaseURL returns the base URL this client talks to, including the /api suffix.
func (c *Client) BaseURL() string { return c.baseURL.String() }

// IsAuthenticated reports whether the client is authenticated.
func (c *Client) IsAuthenticated() bool { return c.authToken != "" }

type requestSpec struct {
	method string
	path   string
	query  url.Values
	body   any
}

func (c *Client) newRequest(ctx context.Context, spec requestSpec) (*http.Request, error) {
	full := c.baseURL.String() + strings.TrimPrefix(spec.path, "/")
	u, err := url.Parse(full)
	if err != nil {
		return nil, newConfigError(fmt.Sprintf("invalid request path %q: %v", spec.path, err))
	}
	if spec.query != nil {
		u.RawQuery = spec.query.Encode()
	}

	var body io.Reader
	if spec.body != nil {
		data, err := json.Marshal(spec.body)
		if err != nil {
			return nil, newConfigError(fmt.Sprintf("encode request body: %v", err))
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, spec.method, u.String(), body)
	if err != nil {
		return nil, newHTTPError(err)
	}

	for name, values := range c.defaultHeaders {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}
	if spec.body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *Client) execute(
	ctx context.Context,
	class rateClass,
	route string,
	conc concurrency,
	spec requestSpec,
) (*http.Response, error) {
	attempt := 0
	for {
		if err := c.limiter.acquire(ctx, class, route); err != nil {
			return nil, asError(err)
		}
		release, err := c.limiter.acquireConcurrency(ctx, conc)
		if err != nil {
			return nil, asError(err)
		}

		req, err := c.newRequest(ctx, spec)
		if err != nil {
			release()
			return nil, err
		}

		resp, err := c.http.Do(req)
		if err != nil {
			release()
			if ctx.Err() != nil {
				return nil, newHTTPError(ctx.Err())
			}
			if isRetryableTransport(err) && attempt < c.maxRetries {
				delay := backoff(attempt, c.retryBaseDelay)
				attempt++
				if serr := sleepCtx(ctx, delay); serr != nil {
					return nil, asError(serr)
				}
				continue
			}
			return nil, newHTTPError(err)
		}

		status := resp.StatusCode
		if status >= 200 && status < 300 {
			release()
			return resp, nil
		}
		if isRetryableStatus(status) && attempt < c.maxRetries {
			delay := retryDelay(resp, attempt, c.retryBaseDelay)
			attempt++
			resp.Body.Close()
			release()
			if serr := sleepCtx(ctx, delay); serr != nil {
				return nil, asError(serr)
			}
			continue
		}
		if status == http.StatusTooManyRequests {
			resp.Body.Close()
			release()
			return nil, newRateLimitedError(attempt + 1)
		}
		apiErr := parseAPIError(resp)
		release()
		return nil, apiErr
	}
}

func fetchJSON[T any](
	c *Client,
	ctx context.Context,
	class rateClass,
	route string,
	conc concurrency,
	spec requestSpec,
) (T, error) {
	var out T
	resp, err := c.execute(ctx, class, route, conc, spec)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return out, newHTTPError(err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return out, newDecodeError(errors.New("empty response body"))
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return out, newDecodeError(err)
	}
	return out, nil
}

func fetchDownload(
	c *Client,
	ctx context.Context,
	class rateClass,
	route string,
	conc concurrency,
	spec requestSpec,
) (*Download, error) {
	resp, err := c.execute(ctx, class, route, conc, spec)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, newHTTPError(err)
	}
	return &Download{ContentType: resp.Header.Get("Content-Type"), Data: data}, nil
}

func fetchOK(
	c *Client,
	ctx context.Context,
	class rateClass,
	route string,
	conc concurrency,
	spec requestSpec,
) error {
	resp, err := c.execute(ctx, class, route, conc, spec)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

func fetchAck(
	c *Client,
	ctx context.Context,
	class rateClass,
	route string,
	conc concurrency,
	spec requestSpec,
) (*OkResponse, error) {
	out, err := fetchJSON[OkResponse](c, ctx, class, route, conc, spec)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func parseAPIError(resp *http.Response) *Error {
	status := resp.StatusCode
	requestID := resp.Header.Get("x-request-id")
	if requestID == "" {
		requestID = resp.Header.Get("request-id")
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	message := ""
	var parsed struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil {
		message = parsed.Error
	} else {
		text := strings.TrimSpace(string(body))
		if text == "" {
			message = statusText(status)
		} else {
			message = text
		}
	}
	return newAPIError(status, message, requestID)
}

func isRetryableStatus(status int) bool {
	switch status {
	case http.StatusRequestTimeout,
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func isRetryableTransport(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	var urlErr *url.Error
	return errors.As(err, &urlErr)
}

func retryDelay(resp *http.Response, attempt int, base time.Duration) time.Duration {
	if value := resp.Header.Get("Retry-After"); value != "" {
		if seconds, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return backoff(attempt, base)
}

func backoff(attempt int, base time.Duration) time.Duration {
	shift := attempt
	if shift > 6 {
		shift = 6
	}
	delay := base * time.Duration(1<<shift)
	if delay <= 0 || delay > 30*time.Second {
		delay = 30 * time.Second
	}
	jitter := time.Duration(randomByte()%250) * time.Millisecond
	return delay + jitter
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// listQuery builds the shared before / include_archived query pairs.
func listQuery(params *ListParams) url.Values {
	query := url.Values{}
	if params == nil {
		return query
	}
	if params.Before != nil {
		query.Set("before", *params.Before)
	}
	if params.IncludeArchived {
		query.Set("include_archived", "true")
	}
	return query
}
