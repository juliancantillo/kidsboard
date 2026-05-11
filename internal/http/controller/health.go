package controller

import (
	"database/sql"
	"net/http"
)

// Health exposes liveness and readiness endpoints for Kubernetes probes.
// Liveness signals process health (always green if we can respond at all);
// readiness signals "I can serve traffic right now" by pinging the DB.
type Health struct {
	db *sql.DB
}

func NewHealth(db *sql.DB) *Health { return &Health{db: db} }

// Live is the liveness probe. Returns 200 if the HTTP handler is responding —
// the kubelet's signal that the container does NOT need to be restarted.
func (c *Health) Live(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// Ready is the readiness probe. Returns 503 when the DB is unreachable so the
// kubelet pulls the Pod out of Service endpoints until the dependency recovers.
func (c *Health) Ready(w http.ResponseWriter, r *http.Request) {
	if err := c.db.PingContext(r.Context()); err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("db unavailable"))
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}
