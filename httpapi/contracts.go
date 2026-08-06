package httpapi

// Wire shapes for /v1 (contract C1). Field order is contract: encoding/json
// preserves struct order (AUDIT.md A5).

type healthResponse struct {
	Status     string `json:"status"`
	GoVersion  string `json:"goVersion"`
	APIVersion string `json:"apiVersion"`
}

type errorEnvelope struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}
