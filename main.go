package main

import (
	"github.com/gorilla/mux"
	"github.com/pkg/browser"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s API - %s\n", r.RemoteAddr, r.URL.String())
		remote, _ := url.Parse(mux.Vars(r)["url"])
		r.URL.Scheme = remote.Scheme
		r.URL.Path = remote.Path
		r.Host = remote.Host
		remote.Path = ""
		r.Header.Del("Referer")
		r.Header.Del("Origin")
		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.ModifyResponse = func(resp *http.Response) error {
			if setCookie := resp.Header.Get("Set-Cookie"); setCookie != "" {
				setCookie = strings.ReplaceAll(setCookie, "Path=/", "Path=/proxy/"+remote.String()+"/")
				resp.Header.Set("Set-Cookie", setCookie)
			}
			return nil
		}
		proxy.ServeHTTP(w, r)
	}
	staticRemote, _ := url.Parse("https://f1vp.netlify.app/")
	staticProxy := httputil.NewSingleHostReverseProxy(staticRemote)
	staticHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s FE* - %s\n", r.RemoteAddr, r.URL.String())
		r.Host = staticRemote.Host
		staticProxy.ServeHTTP(w, r)
	}

	fnRemote, _ := url.Parse("https://fwv-us.deta.dev/")
	fnProxy := httputil.NewSingleHostReverseProxy(fnRemote)
	fnHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s BE* - %s\n", r.RemoteAddr, r.URL.String())
		r.Host = fnRemote.Host
		fnProxy.ServeHTTP(w, r)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{url:https?://.*}", handler)
	r.PathPrefix("/authenticate").HandlerFunc(fnHandler)
	r.PathPrefix("/66571939").HandlerFunc(fnHandler)
	r.PathPrefix("/").HandlerFunc(staticHandler)
	r.SkipClean(true)

	_ = browser.OpenURL("http://localhost:13331/")
	log.Printf("Reverse Proxy Server running at :13331\n")
	http.Handle("/", r)
	panic(http.ListenAndServe(":13331", r))
}
