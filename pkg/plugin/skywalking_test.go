package plugin

import (
	"context"
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
