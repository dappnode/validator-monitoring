package handlers

import "net/http"

func GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	respondOK(w, "Server is running")
}
