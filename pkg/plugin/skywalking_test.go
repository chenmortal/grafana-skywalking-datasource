package plugin

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
)

const wantGenericHealthMessage = "Unable to connect, see Grafana server log for details"

func assertFailedHealthCheck(t *testing.T, result *backend.CheckHealthResult, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("CheckHealth returned error: %v", err)
	}
	if result.Status != backend.HealthStatusError {
		t.Fatalf("expected HealthStatusError, got %v", result.Status)
	}
	if result.Message != wantGenericHealthMessage {
		t.Fatalf("expected generic message %q, got %q", wantGenericHealthMessage, result.Message)
	}
}

// TestCheckHealth_MisconfiguredDatasource_GenericMessage ensures the raw
// getDSInfo error (e.g. "error reading settings: url is empty") is not
// surfaced in the "Save & test" UI.
func TestCheckHealth_MisconfiguredDatasource_GenericMessage(t *testing.T) {
	svc := ProvideService(httpclient.NewProvider())
	pc := emptyURLPluginContext()
	ctx := backend.WithPluginContext(context.Background(), pc)

	result, err := svc.CheckHealth(ctx, &backend.CheckHealthRequest{PluginContext: pc})
	assertFailedHealthCheck(t, result, err)

	if strings.Contains(strings.ToLower(result.Message), "url") {
		t.Fatalf("message leaks underlying error details: %q", result.Message)
	}
}

// TestCheckHealth_UpstreamError_GenericMessage ensures raw connection/GraphQL
// errors from ListLayer (including internal hostnames, IPs, ports) are logged
// server-side instead of being shown in the UI.
func TestCheckHealth_UpstreamError_GenericMessage(t *testing.T) {
	server := newErrorServer(t)
	defer server.Close()

	svc := ProvideService(httpclient.NewProvider())
	pc := backend.PluginContext{
		DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
			URL:      server.URL,
			JSONData: []byte(`{}`),
		},
	}
	ctx := backend.WithPluginContext(context.Background(), pc)

	result, err := svc.CheckHealth(ctx, &backend.CheckHealthRequest{PluginContext: pc})
	assertFailedHealthCheck(t, result, err)

	if strings.Contains(result.Message, "boom") {
		t.Fatalf("message leaks underlying GraphQL error: %q", result.Message)
	}
}

// newV2ProbeServer returns an httptest server that answers the
// hasQueryTracesV2Support probe query with the given raw JSON body. Any other
// query (e.g. listLayers) is answered with a minimal success payload, though
// the v2-unsupported paths return before reaching it.
func newV2ProbeServer(t *testing.T, probeResponse string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		rw.Header().Set("Content-Type", "application/json")
		if strings.Contains(string(body), v2ProbeField) {
			_, _ = rw.Write([]byte(probeResponse))
			return
		}
		_, _ = rw.Write([]byte(`{"data":{"listLayers":["GENERAL"]}}`))
	}))
}

func checkHealthWithV2Enabled(t *testing.T, server *httptest.Server) (*backend.CheckHealthResult, error) {
	t.Helper()
	svc := ProvideService(httpclient.NewProvider())
	pc := backend.PluginContext{
		DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{
			URL:      server.URL,
			JSONData: []byte(`{"interfacev2":true}`),
		},
	}
	ctx := backend.WithPluginContext(context.Background(), pc)
	return svc.CheckHealth(ctx, &backend.CheckHealthRequest{PluginContext: pc})
}

// assertV2UnsupportedGuidance verifies the recovery flow contract: the health
// check must fail with actionable guidance that names the exact setting users
// need to turn off in the (always-toggleable) config editor switch.
func assertV2UnsupportedGuidance(t *testing.T, result *backend.CheckHealthResult, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("CheckHealth returned error: %v", err)
	}
	if result.Status != backend.HealthStatusError {
		t.Fatalf("expected HealthStatusError, got %v", result.Status)
	}
	if result.Message != healthCheckV2UnsupportedMessage {
		t.Fatalf("expected v2-unsupported guidance %q, got %q", healthCheckV2UnsupportedMessage, result.Message)
	}
	if !strings.Contains(result.Message, "Interface Version v2") {
		t.Fatalf("guidance doesn't name the setting to turn off: %q", result.Message)
	}
}

// TestCheckHealth_V2Unsupported_ActionableGuidance covers an OAP server that
// responds to the probe with hasQueryTracesV2Support: false.
func TestCheckHealth_V2Unsupported_ActionableGuidance(t *testing.T) {
	server := newV2ProbeServer(t, `{"data":{"hasQueryTracesV2Support":false}}`)
	defer server.Close()

	result, err := checkHealthWithV2Enabled(t, server)
	assertV2UnsupportedGuidance(t, result, err)
}

// TestCheckHealth_V2ProbeFieldUndefined_ActionableGuidance covers OAP servers
// older than 10.3.0, which don't define the probe field at all: graphql-java
// rejects the query with a FieldUndefined validation error. That must map to
// the same recovery guidance instead of the generic connection error, so users
// on old backends also learn how to recover.
func TestCheckHealth_V2ProbeFieldUndefined_ActionableGuidance(t *testing.T) {
	server := newV2ProbeServer(t, `{"errors":[{"message":"Validation error (FieldUndefined@[1:97]) : Field 'hasQueryTracesV2Support' in type 'Query' is undefined","locations":[{"line":1,"column":97}],"extensions":{"classification":"ValidationError"}}]}`)
	defer server.Close()

	result, err := checkHealthWithV2Enabled(t, server)
	assertV2UnsupportedGuidance(t, result, err)
}

// TestIsV2ProbeFieldUndefined guards the conservative matching: unrelated
// errors (network failures, other GraphQL errors) must keep the generic
// health check message and not be misreported as "v2 unsupported".
func TestIsV2ProbeFieldUndefined(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"graphql-java FieldUndefined", errors.New("Validation error (FieldUndefined@[1:97]) : Field 'hasQueryTracesV2Support' in type 'Query' is undefined"), true},
		{"spec-style cannot query field", errors.New(`Cannot query field "hasQueryTracesV2Support" on type "Query".`), true},
		{"unrelated graphql error", errors.New("boom from OAP"), false},
		{"network error", errors.New("dial tcp 10.0.0.5:12800: connect: connection refused"), false},
		{"undefined without field name", errors.New("some undefined behaviour"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isV2ProbeFieldUndefined(tt.err); got != tt.want {
				t.Errorf("isV2ProbeFieldUndefined(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
