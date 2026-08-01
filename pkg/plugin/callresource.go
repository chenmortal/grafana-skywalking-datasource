package plugin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func (s *Service) registerResourceRoutes() *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("GET /layers", s.withDatasourceHandlerFunc(getListLayerHandler))
	router.HandleFunc("GET /services/{layer}", s.withDatasourceHandlerFunc(queryServicesHandler))
	router.HandleFunc("POST /endpoints", s.withDatasourceHandlerFunc(queryEndpointsHandler))
	router.HandleFunc("POST /instances", s.withDatasourceHandlerFunc(queryInstancesHandler))
	// router.HandleFunc("GET /services/{service}/operations", s.withDatasourceHandlerFunc(getOperationsHandler))
	return router
}

func getListLayerHandler(d *datasourceInfo) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		layers, err := d.SkywalkingClient.ListLayer(r.Context())
		writeResponse(*layers, err, rw, d.SkywalkingClient.logger)
	}
}

func queryServicesHandler(d *datasourceInfo) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		layer := strings.TrimSpace(r.PathValue("layer"))
		services, err := d.SkywalkingClient.QueryServices(r.Context(), layer)
		writeResponse(services, err, rw, d.SkywalkingClient.logger)
	}
}

func queryEndpointsHandler(d *datasourceInfo) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var req queryEndpointsRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeResponse(nil, err, rw, d.SkywalkingClient.logger)
			return
		}

		endpoints, err := d.SkywalkingClient.QueryEndpoints(
			r.Context(),
			req.ServiceId,
			req.Keyword,
			time.UnixMilli(req.FromTime),
			time.UnixMilli(req.ToTime),
			req.Limit)
		writeResponse(*endpoints, err, rw, d.SkywalkingClient.logger)
	}
}

func queryInstancesHandler(d *datasourceInfo) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var req queryInstancesRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeResponse(nil, err, rw, d.SkywalkingClient.logger)
			return
		}

		instances, err := d.SkywalkingClient.QueryInstances(
			r.Context(),
			req.ServiceId,
			time.UnixMilli(req.FromTime),
			time.UnixMilli(req.ToTime))
		writeResponse(instances, err, rw, d.SkywalkingClient.logger)
	}
}

type queryEndpointsRequest struct {
	ServiceId string `json:"serviceId"`
	Keyword   string `json:"keyword"`
	FromTime  int64  `json:"fromTime"`
	ToTime    int64  `json:"toTime"`
	Limit     int    `json:"limit"`
}

type queryInstancesRequest struct {
	ServiceId string `json:"serviceId"`
	FromTime  int64  `json:"fromTime"`
	ToTime    int64  `json:"toTime"`
}

func (s *Service) withDatasourceHandlerFunc(getHandler func(d *datasourceInfo) http.HandlerFunc) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		client, err := s.getDSInfo(r.Context(), backend.PluginConfigFromContext(r.Context()))
		if err != nil {
			writeResponse(nil, errors.New("error getting data source information from context"), rw, client.SkywalkingClient.logger)
			return
		}
		h := getHandler(client)
		h.ServeHTTP(rw, r)
	}
}

func writeResponse(res interface{}, err error, rw http.ResponseWriter, logger log.Logger) {
	if err != nil {
		// This is used for resource calls, we don't need to add actual error message, but we should log it
		logger.Warn("An error occurred while doing a resource call", "error", err)
		http.Error(rw, "An error occurred within the plugin", http.StatusInternalServerError)
		return
	}
	// Response should not be string, but just in case, handle it
	if str, ok := res.(string); ok {
		rw.Header().Set("Content-Type", "text/plain")
		_, _ = rw.Write([]byte(str))
		return
	}
	b, err := json.Marshal(res)
	if err != nil {
		// This is used for resource calls, we don't need to add actual error message, but we should log it
		logger.Warn("An error occurred while processing response from resource call", "error", err)
		http.Error(rw, "An error occurred within the plugin", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	_, _ = rw.Write(b)
}
