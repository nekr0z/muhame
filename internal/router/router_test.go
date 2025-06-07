package router_test

import (
	"bytes"
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/nekr0z/muhame/internal/hash"
	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/nekr0z/muhame/internal/router"
	"github.com/nekr0z/muhame/internal/storage"
)

var testDSN string

func Example() {
	log := zap.NewNop()
	st, _ := storage.New(log.Sugar(), storage.Config{InMemory: true})
	r := router.New(log, st, "", nil)

	srv := httptest.NewServer(r)
	defer srv.Close()

	req, _ := http.NewRequest("POST", srv.URL+"/update/gauge/test/1.2", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	fmt.Printf("Status: %d\n", resp.StatusCode)

	req, _ = http.NewRequest("GET", srv.URL+"/value/gauge/test", nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	result, _ := io.ReadAll(resp.Body)
	fmt.Printf("Value: %s\n", string(result))

	req, _ = http.NewRequest("POST", srv.URL+"/update/", strings.NewReader(`{"id":"test","type":"gauge","value":0.5}`))
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	req, _ = http.NewRequest("GET", srv.URL+"/value/gauge/test", nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	result, _ = io.ReadAll(resp.Body)
	fmt.Printf("Value: %s\n", string(result))

	// Output:
	// Status: 200
	// Value: 1.2
	// Value: 0.5
}

func TestNew_JSONUpdate(t *testing.T) {
	t.Parallel()

	log := zap.NewNop()

	tests := []struct {
		name       string
		in         string
		wantCode   int
		wantRes    string
		wantName   string
		wantMetric metrics.Metric
	}{
		{
			name: "gauge",
			in: `{
				"id": "test",
				"type": "gauge",
				"value": 1.2
			}`,
			wantCode: 200,
			wantRes: `{
				"id": "test",
				"type": "gauge",
				"value": 1.2
			}`,
			wantName:   "test",
			wantMetric: metrics.Gauge(1.2),
		},
		{
			name: "counter",
			in: `{
				"id": "test",
				"type": "counter",
				"delta": 1
				}`,
			wantRes: `{
				"id": "test",
				"type": "counter",
				"delta": 1
				}`,
			wantCode:   200,
			wantName:   "test",
			wantMetric: metrics.Counter(1),
		},
		{
			name: "counter with value",
			in: `{
				"id": "test",
				"type": "counter",
				"value": 1.2
			}`,
			wantCode: 400,
		},
		{
			name: "gauge with delta",
			in: `{
				"id": "test",
				"type": "gauge",
				"delta": 8
			}`,
			wantCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			st := &mockStorage{
				t:    t,
				name: tt.wantName,
				m:    tt.wantMetric,
			}

			r := router.New(log, st, "", nil)

			req := httptest.NewRequest("POST", "/update/", strings.NewReader(tt.in))
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			assert.Equal(t, tt.wantCode, res.Code)

			if tt.wantRes != "" {
				assert.JSONEq(t, tt.wantRes, res.Body.String())
				assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
			}
		})
	}
}

func TestNew_JSONValue(t *testing.T) {
	t.Parallel()

	log := zap.NewNop()

	tests := []struct {
		name     string
		in       string
		m        metrics.Metric
		want     string
		wantCode int
	}{
		{
			name: "gauge",
			in: `{
				"id": "test",
				"type": "gauge"
			}`,
			m: metrics.Gauge(1.2),
			want: `{
				"id": "test",
				"type": "gauge",
				"value": 1.2
			}`,
			wantCode: 200,
		},
		{
			name: "counter",
			in: `{
				"id": "test",
				"type": "counter"
			}`,
			m: metrics.Counter(2),
			want: `{
				"id": "test",
				"type": "counter",
				"delta": 2
			}`,
			wantCode: 200,
		},
		{
			name: "counter",
			in: `{
				"id": "unexpected",
				"type": "counter"
			}`,
			m: metrics.Counter(2),
			want: `{
				"id": "unexpected",
				"type": "counter"
			}`,
			wantCode: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			st := &mockStorage{
				t:    t,
				name: "test",
				m:    tt.m,
			}

			r := router.New(log, st, "", nil)

			req := httptest.NewRequest("POST", "/value/", strings.NewReader(tt.in))
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			assert.Equal(t, tt.wantCode, res.Code)

			assert.JSONEq(t, tt.want, res.Body.String())
		})
	}
}

func TestNew_Root(t *testing.T) {
	t.Parallel()

	log := zap.NewNop()
	st := &mockStorage{
		t: t,
	}

	r := router.New(log, st, "", nil)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)
	assert.Contains(t, res.Header().Values("Content-Type"), "text/html")
}

func TestNew_Ping(t *testing.T) {
	t.Parallel()

	log := zap.NewNop()

	tests := []struct {
		name  string
		dsn   string
		close bool
		want  int
	}{
		{
			name: "good db",
			dsn:  testDSN,
			want: http.StatusOK,
		},
		{
			name:  "closed db",
			dsn:   testDSN,
			close: true,
			want:  http.StatusInternalServerError,
		},
		{
			name: "no db",
			dsn:  "",
			want: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			st, err := storage.New(log.Sugar(), storage.Config{
				DatabaseDSN: tt.dsn,
			})
			require.NoError(t, err)

			if tt.close {
				st.Close()
			} else {
				t.Cleanup(st.Close)
			}

			r := router.New(log, st, "", nil)
			req := httptest.NewRequest("GET", "/ping", nil)
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			assert.Equal(t, tt.want, res.Code)
		})
	}
}

func TestNew_Updates(t *testing.T) {
	t.Parallel()

	log := zap.NewNop()

	st, err := storage.New(log.Sugar(), storage.Config{
		DatabaseDSN: testDSN,
	})
	require.NoError(t, err)

	in := `[
{
	"id": "gauge_1",
	"type": "gauge",
	"value": 1.2
},
{
	"id": "gauge_2",
	"type": "gauge",
	"value": 2.56
},
{
	"id": "counter_1",
	"type": "counter",
	"delta": 43
},
{
	"id": "counter_1",
	"type": "counter",
	"delta": 64
}
]`

	r := router.New(log, st, "", nil)

	req := httptest.NewRequest("POST", "/updates/", strings.NewReader(in))
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code, res.Body.String())

	req = httptest.NewRequest("GET", "/", nil)
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	body, _ := io.ReadAll(res.Body)

	bs := string(body)
	assert.Contains(t, bs, "gauge_1")
	assert.Contains(t, bs, "gauge_2")
	assert.Contains(t, bs, "counter_1")
	assert.Contains(t, bs, "1.2")
	assert.Contains(t, bs, "2.56")
	assert.Contains(t, bs, "107")
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	dbName := "metrics"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
			wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		panic(err)
	}

	testDSN, err = postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	m.Run()

	err = testcontainers.TerminateContainer(postgresContainer)
	if err != nil {
		panic(err)
	}
}

func TestNew_Signature(t *testing.T) {
	in := `{"id":"test","type":"counter","delta":1}`
	key := "testkey"
	sig := hash.Signature([]byte(in), key)

	st := &mockStorage{
		t:    t,
		name: "test",
		m:    metrics.Counter(1),
	}

	log := zap.NewNop()

	r := router.New(log, st, "testkey", &rsa.PrivateKey{})

	body := bytes.NewBufferString(in)

	req := httptest.NewRequest(http.MethodPost, "/update/", body)
	req.Header.Add(hash.Header, sig)
	req.Header.Add("Content-Type", "application/json")
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	result := res.Result()
	defer func() {
		err := result.Body.Close()
		assert.NoError(t, err)
	}()

	b, err := io.ReadAll(result.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, result.StatusCode, string(b))
	assert.Contains(t, result.Header, hash.Header)
	assert.Equal(t, sig, result.Header.Get(hash.Header))
}

var _ storage.Storage = &mockStorage{}

type mockStorage struct {
	t    *testing.T
	name string
	m    metrics.Metric
}

func (m mockStorage) Update(_ context.Context, named metrics.Named) error {
	m.t.Helper()
	assert.Equal(m.t, m.name, named.Name)
	assert.Equal(m.t, m.m, named.Metric)
	return nil
}

func (m mockStorage) Get(_ context.Context, metricType, name string) (metrics.Metric, error) {
	m.t.Helper()

	if name != m.name {
		return nil, storage.ErrMetricNotFound
	}

	return m.m, nil
}

func (m mockStorage) List(_ context.Context) ([]metrics.Named, error) {
	m.t.Helper()
	return nil, nil
}

func (m mockStorage) Close() {
	m.t.Helper()
}
