package handlers

import (
	"net/http"
)

func StaticsHandler() http.Handler {
	fs := http.FileServer(http.Dir("./static"))
	return http.StripPrefix("/", fs)
}