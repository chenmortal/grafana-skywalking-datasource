package plugin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// newErrorServer returns an httptest server that always answers with a
// GraphQL-level error. genqlient clients return (nil, err) for such responses,
// which is the situation when the OAP server rejects a query (e.g. a freshly
// installed OAP with no matching trace data, or an unreachable endpoint).
func newErrorServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"errors":[{"message":"boom from OAP"}]}`))
	}))
}

func newTestDSInfo(t *testing.T, server *httptest.Server, v2 bool) *datasourceInfo {
	t.Helper()
	client, err := New(server.Client(), log.DefaultLogger, server.URL)
	if err != nil {
		t.Fatalf("failed to create skywalking client: %v", err)
	}
	return &datasourceInfo{
		SkywalkingClient: client,
		SourceSettings:   backend.DataSourceInstanceSettings{UID: "test-uid", Name: "test-ds"},
		PluginSettings:   PluginSettings{V2: v2},
	}
}

// TestQueryData_UpstreamGraphQLError guards the fix for the backend crash:
// the transform functions used to be called BEFORE the error of the preceding
// GraphQL call was checked, causing a nil pointer dereference (panic) that
// takes down the whole plugin backend process.
func TestQueryData_UpstreamGraphQLError(t *testing.T) {
	now := time.Now()
	searchQuery := []byte(`{"queryType":"search","layer":"GENERAL","condition":{"traceState":"ALL","queryOrder":"BY_START_TIME","paging":{"pageNum":1,"pageSize":20}}}`)
	traceQuery := []byte(`{"queryType":"","query":"abc123traceid"}`)

	tests := []struct {
		name      string
		v2        bool
		queryJSON []byte
	}{
		{name: "search basic (v1)", v2: false, queryJSON: searchQuery},
		{name: "search v2", v2: true, queryJSON: searchQuery},
		{name: "trace by id (v1)", v2: false, queryJSON: traceQuery},
		{name: "trace by id v2", v2: true, queryJSON: traceQuery},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newErrorServer(t)
			defer server.Close()
			dsInfo := newTestDSInfo(t, server, tt.v2)

			req := &backend.QueryDataRequest{
				Queries: []backend.DataQuery{
					{
						RefID:     "A",
						TimeRange: backend.TimeRange{From: now.Add(-time.Hour), To: now},
						JSON:      tt.queryJSON,
					},
				},
			}

			// Must not panic; the upstream error must surface as an error response.
			resp, err := queryData(context.Background(), dsInfo, req)
			if err != nil {
				t.Fatalf("queryData returned error: %v", err)
			}
			dr, ok := resp.Responses["A"]
			if !ok {
				t.Fatal("missing response for refID A")
			}
			if dr.Error == nil {
				t.Fatal("expected error response for failed upstream query, got none")
			}
			if len(dr.Frames) != 0 {
				t.Fatalf("expected no frames on error, got %d", len(dr.Frames))
			}
		})
	}
}
