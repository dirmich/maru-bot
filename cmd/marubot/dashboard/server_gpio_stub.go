//go:build !linux || (!arm && !arm64)

package dashboard

import "net/http"

// isRPi returns false on non-RPi platforms
func isRPi() bool { return false }

func (s *Server) registerGPIORoutes(mux *http.ServeMux) {
	// GPIO not supported on this platform
	notSupported := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"GPIO not supported on this platform"}`, http.StatusNotImplemented)
	}
	mux.HandleFunc("/api/gpio", notSupported)
	mux.HandleFunc("/api/gpio/toggle", notSupported)
}
