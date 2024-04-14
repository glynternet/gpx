package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	ggpx "github.com/glynternet/gpx/pkg/gpx"
	"github.com/tkrajina/gpxgo/gpx"
)

func HandleElevation(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "content-type")
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed: "+http.MethodPost+" only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	content, err := gpx.Parse(r.Body)
	if err != nil {
		_ = r.Body.Close()
		http.Error(w, fmt.Sprintf("parsing body as gpx: %s", err), http.StatusBadRequest)
		return
	}
	if err := r.Body.Close(); err != nil {
		http.Error(w, fmt.Sprintf("closing request body: %s", err), http.StatusBadRequest)
		return
	}
	profile, err := ggpx.Profile(*content)
	if err != nil {
		http.Error(w, fmt.Sprintf("calculating elevation profile: %s", err), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(profile); err != nil {
		http.Error(w, fmt.Sprintf("json encoding elevation profile: %s", err), http.StatusInternalServerError)
		return
	}
}
