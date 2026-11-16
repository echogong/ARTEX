package server

import (
	"net/http"
	"runtime"
)

// onlineUpdateDisabled is the only response the former self-update routes
// return. The check, stage, and apply path is compiled out: these handlers do
// not contact GitHub, do not read an update token, and do not replace the
// running binary. Upgrades are a manual git pull and rebuild.
const onlineUpdateDisabled = "online update disabled; upgrade manually"

// updateCheck reports the running version and that online update is disabled.
// It does not look up a release.
func (s *Server) updateCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"current":       BuildVersion,
		"online_update": false,
		"message":       onlineUpdateDisabled,
		"os":            runtime.GOOS,
		"arch":          runtime.GOARCH,
	})
}

// updateApply refuses to download or stage a release.
func (s *Server) updateApply(w http.ResponseWriter, r *http.Request) {
	writeErr(w, http.StatusGone, onlineUpdateDisabled)
}

// updateRollback refuses to swap in a previously staged binary.
func (s *Server) updateRollback(w http.ResponseWriter, r *http.Request) {
	writeErr(w, http.StatusGone, onlineUpdateDisabled)
}

// updateStream refuses the old progress stream. There is no download to report.
func (s *Server) updateStream(w http.ResponseWriter, r *http.Request) {
	writeErr(w, http.StatusGone, onlineUpdateDisabled)
}
