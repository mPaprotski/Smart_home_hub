package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coder/websocket"
	"github.com/mPaprotski/Smart_home_hub/internal/auth"
	"github.com/mPaprotski/Smart_home_hub/internal/storage"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func securedFixture(t *testing.T) (*Server, *auth.Access, *httptest.Server, *http.Client) {
	t.Helper()
	db, err := storage.Open(context.Background(), storage.Options{Path: filepath.Join(t.TempDir(), "private", "hub.db"), BusyTimeout: time.Second, StartupTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	service, err := auth.New(context.Background(), db, auth.NewHasher())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Bootstrap(context.Background(), "admin", []byte("Synthetic password sentinel for HTTPS")); err != nil {
		t.Fatal(err)
	}
	s, h := fixture("sqlite")
	s.cfg.Auth.Enabled = true
	s.SetAuthenticator(service)
	a := auth.NewAccess(db, service)
	s.SetAccess(a)
	h.Set("storage", "READY")
	s.cfg.Auth.PublicOrigin = "https://placeholder"
	ts := httptest.NewUnstartedServer(s.Handler())
	ts.StartTLS()
	s.cfg.Auth.PublicOrigin = ts.URL
	client := ts.Client()
	jar, _ := cookiejar.New(nil)
	client.Jar = jar
	t.Cleanup(func() { a.Close(); ts.Close() })
	return s, a, ts, client
}
func apiCall(t *testing.T, client *http.Client, method, uri, origin, csrf, bearer string, body any) (int, http.Header, map[string]any) {
	t.Helper()
	var input io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		input = strings.NewReader(string(b))
	}
	r, err := http.NewRequest(method, uri, input)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	r.Header.Set("X-Hub-Request", "1")
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	if csrf != "" {
		r.Header.Set("X-CSRF-Token", csrf)
	}
	if bearer != "" {
		r.Header.Set("Authorization", "Bearer "+bearer)
	}
	response, err := client.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var out map[string]any
	raw, _ := io.ReadAll(response.Body)
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	return response.StatusCode, response.Header, out
}
func login(t *testing.T, client *http.Client, base string) (csrf, user string) {
	t.Helper()
	code, headers, body := apiCall(t, client, "POST", base+"/api/v1/auth/login", base, "", "", map[string]string{"username": "admin", "password": "Synthetic password sentinel for HTTPS"})
	if code != 200 {
		t.Fatal(code, body)
	}
	cookie := headers.Get("Set-Cookie")
	for _, flag := range []string{"__Host-shh_session=", "Secure", "HttpOnly", "SameSite=Lax", "Path=/"} {
		if !strings.Contains(cookie, flag) {
			t.Fatal("cookie flag", flag)
		}
	}
	if strings.Contains(cookie, "Domain=") {
		t.Fatal("host cookie domain")
	}
	return body["csrf_token"].(string), body["user_id"].(string)
}
func TestHTTPSLoginCSRFMutationLogoutAndBearer(t *testing.T) {
	_, _, ts, client := securedFixture(t)
	csrf, user := login(t, client, ts.URL)
	code, _, _ := apiCall(t, client, "GET", ts.URL+"/api/v1/auth/me", "", "", "", nil)
	if code != 200 {
		t.Fatal(code)
	}
	body := map[string]any{"value": "Europe/Minsk", "expected_revision": 0, "command_id": "one"}
	for _, origin := range []string{"", "null", "https://evil.example"} {
		code, _, _ := apiCall(t, client, "PUT", ts.URL+"/api/v1/users/me/settings/timezone", origin, csrf, "", body)
		if code != 403 {
			t.Fatal("origin bypass", origin, code)
		}
	}
	code, _, _ = apiCall(t, client, "PUT", ts.URL+"/api/v1/users/me/settings/timezone", ts.URL, "bad", "", body)
	if code != 403 {
		t.Fatal("CSRF bypass", code)
	}
	code, _, out := apiCall(t, client, "PUT", ts.URL+"/api/v1/users/me/settings/timezone", ts.URL, csrf, "", body)
	if code != 200 || out["revision"] != float64(1) {
		t.Fatal(code, out)
	}
	code, _, out = apiCall(t, client, "POST", ts.URL+"/api/v1/auth/tokens", ts.URL, csrf, "", map[string]any{"name": "fixture", "scopes": []string{"settings:write"}, "bindings": []string{"user:" + user}})
	if code != 201 {
		t.Fatal(code, out)
	}
	token := out["token"].(string)
	id := out["metadata"].(map[string]any)["id"].(string)
	bearerClient := *client
	bearerClient.Jar = nil
	body["expected_revision"] = 1
	body["command_id"] = "machine"
	body["value"] = "UTC"
	code, _, out = apiCall(t, &bearerClient, "PUT", ts.URL+"/api/v1/users/me/settings/timezone", "", "", token, body)
	if code != 200 {
		t.Fatal(code, out)
	}
	code, _, _ = apiCall(t, client, "GET", ts.URL+"/api/v1/auth/me", "", "", token, nil)
	if code != 400 {
		t.Fatal("ambiguous credentials", code)
	}
	code, _, _ = apiCall(t, client, "POST", ts.URL+"/api/v1/auth/tokens/revoke", ts.URL, csrf, "", map[string]string{"id": id})
	if code != 204 {
		t.Fatal(code)
	}
	code, _, _ = apiCall(t, &bearerClient, "PUT", ts.URL+"/api/v1/users/me/settings/timezone", "", "", token, body)
	if code != 401 {
		t.Fatal("revoked retry", code)
	}
	code, headers, _ := apiCall(t, client, "POST", ts.URL+"/api/v1/auth/logout", ts.URL, csrf, "", nil)
	if code != 204 || !strings.Contains(headers.Get("Set-Cookie"), "Max-Age=0") {
		t.Fatal(code, headers)
	}
	code, _, _ = apiCall(t, client, "GET", ts.URL+"/api/v1/auth/me", "", "", "", nil)
	if code != 401 {
		t.Fatal(code)
	}
}
func TestWebsocketActualUpgradeRevokeAndCleanup(t *testing.T) {
	s, a, ts, client := securedFixture(t)
	csrf, _ := login(t, client, ts.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	headers := http.Header{"Origin": []string{ts.URL}}
	conn, _, err := websocket.Dial(ctx, "wss"+strings.TrimPrefix(ts.URL, "https")+"/ws/events", &websocket.DialOptions{HTTPClient: client, HTTPHeader: headers})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	_, data, err := conn.Read(ctx)
	if err != nil || !strings.Contains(string(data), "auth_only") {
		t.Fatal(string(data), err)
	}
	code, _, _ := apiCall(t, client, "POST", ts.URL+"/api/v1/auth/revoke-all", ts.URL, csrf, "", nil)
	if code != 204 {
		t.Fatal(code)
	}
	if _, _, err = conn.Read(ctx); err == nil {
		t.Fatal("WS survived revoke")
	}
	deadline := time.Now().Add(time.Second)
	for a.Connections() != 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if a.Connections() != 0 {
		t.Fatal("registry leaked")
	}
	other := *client
	other.Jar = nil
	if connection, response, err := websocket.Dial(ctx, "wss"+strings.TrimPrefix(ts.URL, "https")+"/ws/events", &websocket.DialOptions{HTTPClient: &other, HTTPHeader: headers}); err == nil {
		connection.CloseNow()
		t.Fatal("anonymous upgrade")
	} else if response.StatusCode != 401 {
		t.Fatal(response.StatusCode)
	}
	_ = s
}
func TestProxyTLSBoundaryAndSecureCookieJar(t *testing.T) {
	s, _, _, _ := securedFixture(t)
	plain := httptest.NewServer(s.Handler())
	defer plain.Close()
	target, _ := url.Parse(plain.URL)
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = http.DefaultTransport
	old := proxy.Director
	proxy.Director = func(r *http.Request) {
		old(r)
		r.Header.Set("X-Forwarded-Proto", "https")
		r.Header.Del("Forwarded")
		r.Header.Del("X-Forwarded-For")
	}
	front := httptest.NewTLSServer(proxy)
	defer front.Close()
	s.cfg.Auth.PublicOrigin = front.URL
	s.cfg.Auth.TrustedProxies = []string{"127.0.0.1/32"}
	frontClient := front.Client()
	frontClient.Jar, _ = cookiejar.New(nil)
	code, headers, _ := apiCall(t, frontClient, "POST", front.URL+"/api/v1/auth/login", front.URL, "", "", map[string]string{"username": "admin", "password": "Synthetic password sentinel for HTTPS"})
	if code != 200 {
		t.Fatal(code)
	}
	frontURL, _ := url.Parse(front.URL)
	// Go1.27/browser jars intentionally treat loopback as a secure context even
	// over HTTP. Test transport flags on a non-loopback logical host instead.
	logical, _ := url.Parse("https://hub.example.test")
	checkJar, _ := cookiejar.New(nil)
	checkJar.SetCookies(logical, (&http.Response{Header: headers}).Cookies())
	if len(checkJar.Cookies(logical)) == 0 {
		t.Fatal("secure cookie missing")
	}
	logical.Scheme = "http"
	if len(checkJar.Cookies(logical)) != 0 {
		t.Fatal("Secure cookie sent to non-loopback HTTP")
	}

	r := httptest.NewRequest("POST", "http://private/api/v1/auth/login", nil)
	r.Host = frontURL.Host
	r.RemoteAddr = net.JoinHostPort("192.0.2.1", "1234")
	r.Header.Set("X-Forwarded-Proto", "https")
	r.Header.Set("Origin", front.URL)
	if s.secure(r) {
		t.Fatal("untrusted forwarded proto")
	}
	r.RemoteAddr = "127.0.0.1:1234"
	if !s.secure(r) {
		t.Fatal("explicit proxy not recognized")
	}
	_ = fmt.Sprint(frontURL)
}

func TestWebsocketIdleExpiryWithoutClientTraffic(t *testing.T) {
	s, a, ts, client := securedFixture(t)
	var clock atomic.Int64
	clock.Store(time.Now().Unix())
	a.Now = func() time.Time { return time.Unix(clock.Load(), 0) }
	s.wsRecheckInterval = 10 * time.Millisecond
	login(t, client, ts.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "wss"+strings.TrimPrefix(ts.URL, "https")+"/ws/events", &websocket.DialOptions{HTTPClient: client, HTTPHeader: http.Header{"Origin": []string{ts.URL}}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	if _, _, err = conn.Read(ctx); err != nil {
		t.Fatal(err)
	}
	clock.Add(int64(storage.SessionIdle / time.Second))
	if _, _, err = conn.Read(ctx); err == nil {
		t.Fatal("idle connection remained authorized")
	}
}

func TestDroppedNotificationKeepsDurableMutation(t *testing.T) {
	s, a, ts, client := securedFixture(t)
	csrf, user := login(t, client, ts.URL)
	s.bus.Close()
	code, _, out := apiCall(t, client, "PUT", ts.URL+"/api/v1/users/me/settings/timezone", ts.URL, csrf, "", map[string]any{"value": "UTC", "expected_revision": 0, "command_id": "bus-closed"})
	if code != 200 {
		t.Fatal(code, out)
	}
	setting, err := a.Store.GetSetting(context.Background(), user, "timezone")
	if err != nil || setting.Revision != 1 {
		t.Fatal(setting, err)
	}
	if s.notificationFailures.Load() != 1 {
		t.Fatal("drop not counted")
	}
	raw := client.Jar.Cookies(mustURL(t, ts.URL))[0].Value
	c, err := a.Verify(context.Background(), raw, "", "peer")
	if err != nil {
		t.Fatal(err)
	}
	audit, err := a.Store.ListAudit(context.Background(), c, 0, a.Now())
	if err != nil || len(audit) != 2 {
		t.Fatal(audit, err)
	}
}
func mustURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, e := url.Parse(raw)
	if e != nil {
		t.Fatal(e)
	}
	return u
}
