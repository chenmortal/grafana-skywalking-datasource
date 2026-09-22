package plugin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
)

func emptyURLPluginContext() backend.PluginContext {
	return backend.PluginContext{
		DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{URL: ""},
	}
}

// TestWithDatasourceHandlerFunc_MisconfiguredDatasource guards the nil-client
// panic: the error branch used to read client.SkywalkingClient.logger while
// getDSInfo guarantees client == nil on every error path.
func TestWithDatasourceHandlerFunc_MisconfiguredDatasource(t *testing.T) {
	svc := ProvideService(httpclient.NewProvider())
	handler := svc.withDatasourceHandlerFunc(getListLayerHandler)

	req := httptest.NewRequest(http.MethodGet, "/layers", nil)
	req = req.WithContext(backend.WithPluginContext(req.Context(), emptyURLPluginContext()))
	rec := httptest.NewRecorder()

	// Must not panic before the fix.
	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "An error occurred within the plugin") {
		t.Fatalf("expected generic error body, got %q", rec.Body.String())
	}
}

// TestGetListLayerHandler_UpstreamError guards the dereference-before-check
// pattern writeResponse(*layers, err, ...) on the resource call sites.
func TestGetListLayerHandler_UpstreamError(t *testing.T) {
	server := newErrorServer(t)
	defer server.Close()
	dsInfo := newTestDSInfo(t, server, false)

	handler := getListLayerHandler(dsInfo)
	req := httptest.NewRequest(http.MethodGet, "/layers", nil)
	rec := httptest.NewRecorder()

	// Must not panic before the fix.
	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}

// TestQueryEndpointsHandler_UpstreamError guards the same dereference pattern
// for writeResponse(*endpoints, err, ...).
func TestQueryEndpointsHandler_UpstreamError(t *testing.T) {
	server := newErrorServer(t)
	defer server.Close()
	dsInfo := newTestDSInfo(t, server, false)

	handler := queryEndpointsHandler(dsInfo)
	body := `{"serviceId":"svc==","keyword":"","fromTime":1700000000000,"toTime":1700003600000,"limit":10}`
	req := httptest.NewRequest(http.MethodPost, "/endpoints", strings.NewReader(body))
	rec := httptest.NewRecorder()

	// Must not panic before the fix.
	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}

// TestGetListLayerHandler_Success_JSONUnchanged pins the success-path output:
// passing the pointer into writeResponse must serialize identically to the
// previous dereferenced value.
func TestGetListLayerHandler_Success_JSONUnchanged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write([]byte(`{"data":{"layers":["GENERAL","MESH"]}}`))
	}))
	defer server.Close()
	dsInfo := newTestDSInfo(t, server, false)

	handler := getListLayerHandler(dsInfo)
	req := httptest.NewRequest(http.MethodGet, "/layers", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var got map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("response is not valid JSON: %v (body: %s)", err, rec.Body.String())
	}
	wantJSON, _ := json.Marshal(map[string]interface{}{"layers": []interface{}{"GENERAL", "MESH"}})
	gotJSON, _ := json.Marshal(got)
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("unexpected body: got %s, want %s", gotJSON, wantJSON)
	}
}
