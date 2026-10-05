package gametorch

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestClient(t *testing.T, serverURL string, opts ...Option) *Client {
	t.Helper()
	base := []Option{WithBaseURL(serverURL)}
	base = append(base, opts...)
	client, err := NewClient(base...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

func TestClientSendsAuthAndHeaders(t *testing.T) {
	var gotAuth, gotAccept, gotUserAgent, gotContentType, gotMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		gotUserAgent = r.Header.Get("User-Agent")
		gotContentType = r.Header.Get("Content-Type")
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"p1","name":"My Game","slug":"my-game","created_at":"2026-01-01T00:00:00Z"}`)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/api", WithAPIKey("gt2_test"), WithRateLimit(false))
	project, err := client.CreateProject(context.Background(), "My Game")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if project.ID != "p1" || project.Slug != "my-game" {
		t.Errorf("unexpected project: %+v", project)
	}
	if gotAuth != "Bearer gt2_test" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotAccept != "application/json" {
		t.Errorf("Accept = %q", gotAccept)
	}
	if !strings.HasPrefix(gotUserAgent, "gametorch-go/") {
		t.Errorf("User-Agent = %q", gotUserAgent)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q", gotContentType)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q", gotMethod)
	}
}

func TestClientBaseURLNormalization(t *testing.T) {
	client, err := NewClient(WithBaseURL("http://localhost:8300/api"))
	if err != nil {
		t.Fatal(err)
	}
	if client.BaseURL() != "http://localhost:8300/api/" {
		t.Errorf("BaseURL = %q", client.BaseURL())
	}
	if client.IsAuthenticated() {
		t.Error("client should not be authenticated")
	}

	if _, err := NewClient(WithBaseURL("://bad")); err == nil {
		t.Error("expected invalid base URL error")
	} else if e := asError(err); e.Kind() != ErrorKindInvalidBaseURL {
		t.Errorf("kind = %q", e.Kind())
	}
}

func TestClientFromEnv(t *testing.T) {
	t.Setenv(EnvAPIKey, "gt2_env")
	t.Setenv(EnvBaseURL, "http://localhost:8300/api")
	client, err := NewClientFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if !client.IsAuthenticated() {
		t.Error("expected env API key to authenticate")
	}
	if client.BaseURL() != "http://localhost:8300/api/" {
		t.Errorf("BaseURL = %q", client.BaseURL())
	}
}

func TestClientRetriesOnServerError(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = io.WriteString(w, `{"projects":[]}`)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL,
		WithRateLimit(false),
		WithMaxRetries(3),
		WithRetryBaseDelay(time.Millisecond),
	)
	if _, err := client.ListProjects(context.Background()); err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestClientHonorsRetryAfter(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = io.WriteString(w, `{"projects":[]}`)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL,
		WithRateLimit(false),
		WithMaxRetries(2),
		WithRetryBaseDelay(time.Millisecond),
	)
	start := time.Now()
	if _, err := client.ListProjects(context.Background()); err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if elapsed := time.Since(start); elapsed < 900*time.Millisecond {
		t.Errorf("retry did not honor Retry-After: elapsed %v", elapsed)
	}
}

func TestClientRateLimitedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, WithRateLimit(false), WithMaxRetries(0))
	_, err := client.ListProjects(context.Background())
	if err == nil {
		t.Fatal("expected rate limit error")
	}
	e := asError(err)
	if !e.IsRateLimited() {
		t.Errorf("IsRateLimited = false (kind %q)", e.Kind())
	}
	if e.Attempts() != 1 {
		t.Errorf("Attempts = %d, want 1", e.Attempts())
	}
}

func TestClientAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-request-id", "req-123")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":"no such asset"}`)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, WithRateLimit(false))
	_, err := client.GetAsset(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected API error")
	}
	e := asError(err)
	if !e.IsNotFound() {
		t.Error("expected IsNotFound")
	}
	if e.APIMessage() != "no such asset" {
		t.Errorf("message = %q", e.APIMessage())
	}
	if e.RequestID() != "req-123" {
		t.Errorf("request id = %q", e.RequestID())
	}
	if e.StatusCode() != http.StatusNotFound {
		t.Errorf("status = %d", e.StatusCode())
	}
}

func TestClientDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "not json")
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, WithRateLimit(false))
	_, err := client.ListProjects(context.Background())
	if err == nil {
		t.Fatal("expected decode error")
	}
	if e := asError(err); e.Kind() != ErrorKindDecode {
		t.Errorf("kind = %q", e.Kind())
	}
}

func TestClientDownload(t *testing.T) {
	payload := []byte{0x89, 'P', 'N', 'G'}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, WithRateLimit(false))
	download, err := client.AssetContent(context.Background(), "a1")
	if err != nil {
		t.Fatal(err)
	}
	if download.ContentType != "image/png" {
		t.Errorf("content type = %q", download.ContentType)
	}
	if download.Len() != len(payload) || download.IsEmpty() {
		t.Errorf("payload len = %d", download.Len())
	}
}

func TestClientListQuery(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"generations":[],"next_cursor":null,"total":0}`)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, WithRateLimit(false))
	params := NewListParams().WithBefore("cursor-1").WithIncludeArchived(true)
	if _, err := client.ListGenerations(context.Background(), "p1", params); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotQuery, "before=cursor-1") || !strings.Contains(gotQuery, "include_archived=true") {
		t.Errorf("query = %q", gotQuery)
	}
}

func TestClientRequestIDHeader(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = io.WriteString(w, `{"id":"j1","status":"queued","created":true,"reserved_credits":"5.000000000000"}`)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, WithRateLimit(false))
	job, err := client.GenerateSprite("p1").
		WithPrompt("a fox").
		WithImageModel("m1").
		WithRequestID("11111111-1111-4111-8111-111111111111").
		Send(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if job.ID != "j1" || !job.Created {
		t.Errorf("job = %+v", job)
	}
	if body["request_id"] != "11111111-1111-4111-8111-111111111111" {
		t.Errorf("request_id = %v", body["request_id"])
	}
	if body["mode"] != "single" {
		t.Errorf("mode = %v", body["mode"])
	}
}

func TestGenerateSpriteRequiresFields(t *testing.T) {
	client, err := NewClient(WithBaseURL("http://localhost:8300/api"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GenerateSprite("p1").Send(context.Background()); err == nil {
		t.Error("expected prompt error")
	}
	if _, err := client.GenerateSprite("p1").WithPrompt("x").Send(context.Background()); err == nil {
		t.Error("expected image model error")
	}
}

func TestPaginator(t *testing.T) {
	pages := []Page[int]{
		{Items: []int{1, 2}, NextCursor: Ptr("c1"), Total: 3},
		{Items: []int{3}, Total: 3},
	}
	call := 0
	p := newPaginator(func(ctx context.Context, cursor *string) (Page[int], error) {
		page := pages[call]
		call++
		return page, nil
	})

	var got []int
	for {
		item, ok, err := p.Next(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			break
		}
		got = append(got, item)
	}
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Errorf("items = %v", got)
	}
	if total, ok := p.Total(); !ok || total != 3 {
		t.Errorf("total = %d, ok = %v", total, ok)
	}
}

func TestPaginatorNextPage(t *testing.T) {
	pages := []Page[string]{
		{Items: []string{"a"}, NextCursor: Ptr("c1"), Total: 2},
		{Items: []string{"b"}, Total: 2},
	}
	call := 0
	p := newPaginator(func(ctx context.Context, cursor *string) (Page[string], error) {
		page := pages[call]
		call++
		return page, nil
	})

	first, ok, err := p.NextPage(context.Background())
	if err != nil || !ok || len(first) != 1 {
		t.Fatalf("first page: %v %v %v", first, ok, err)
	}
	second, ok, err := p.NextPage(context.Background())
	if err != nil || !ok || len(second) != 1 {
		t.Fatalf("second page: %v %v %v", second, ok, err)
	}
	if _, ok, _ := p.NextPage(context.Background()); ok {
		t.Error("expected exhausted paginator")
	}
}

func TestHealthTrimsBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok\n")
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, WithRateLimit(false))
	status, err := client.Health(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status != "ok" {
		t.Errorf("health = %q", status)
	}
}
