// +build linux

package main

import (
	"net/http"
	"net/http/pprof"
)

func pprofHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hi"))
}

func pprofdump() {
	r := http.NewServeMux()
	r.HandleFunc("/", pprofHandler)

	// Регистрация pprof-обработчиков
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	http.ListenAndServe(":8080", r)

}
