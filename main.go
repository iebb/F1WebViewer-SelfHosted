package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pkg/browser"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
)

var builtInSSL = false
var selfSSL = false
var listenAddr = "localhost:13331"
var login = ""
var password = ""

func init() {
	flag.BoolVar(&builtInSSL, "ssl", false,
		"built-in ssl support. this requires a public IP, and port 80 being available in order to fetch the certs",
	)
	flag.BoolVar(&selfSSL, "self-sign", false,
		"self-signed ssl support. would display a warning on browsers",
	)
	flag.StringVar(&listenAddr, "listen", "localhost:13331", "which IP and port to listen?")
	flag.StringVar(&login, "login", "", "server-side login")
	flag.StringVar(&password, "password", "", "server-side password")
}

func main() {
	flag.Parse()

	stripDomain := regexp.MustCompile(`Domain=[^;]+; `)
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s API - %s\n", r.RemoteAddr, r.URL.String())
		varUrl := mux.Vars(r)["url"]
		pathCleaned := !strings.Contains(varUrl, "://")
		if pathCleaned {
			varUrl = strings.ReplaceAll(varUrl, ":/", "://")
		}
		remote, _ := url.Parse(varUrl)
		r.URL.Scheme = remote.Scheme
		r.URL.Path = remote.Path
		r.Host = remote.Host
		remote.Path = ""
		r.Header.Del("Referer")
		r.Header.Del("Origin")
		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.ModifyResponse = func(resp *http.Response) error {
			if setCookie := resp.Header.Get("Set-Cookie"); setCookie != "" {
				remotePath := remote.String()
				if pathCleaned {
					remotePath = strings.ReplaceAll(remotePath, "://", ":/")
				}
				setCookie = strings.ReplaceAll(setCookie, "Path=/", "Path=/proxy/"+remotePath+"/")
				setCookie = strings.ReplaceAll(setCookie, ";SameSite=None;Secure", "")
				setCookie = stripDomain.ReplaceAllString(setCookie, "")
				resp.Header.Set("Set-Cookie", setCookie)
			}
			return nil
		}
		proxy.ServeHTTP(w, r)
	}

	staticRemote, _ := url.Parse("https://f1vp.netlify.app")
	staticProxy := httputil.NewSingleHostReverseProxy(staticRemote)
	if login != "" && password != "" {
		staticProxy.ModifyResponse = serverSideLoginRewriter
	}
	staticHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s FE* - %s\n", r.RemoteAddr, r.URL.String())
		r.Host = staticRemote.Host
		staticProxy.ServeHTTP(w, r)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{url:https?://?.*}", handler)
	r.PathPrefix("/authenticate").HandlerFunc(handleLogin)
	r.PathPrefix("/").HandlerFunc(staticHandler)
	r.SkipClean(true)

	if builtInSSL {
		certManager := autocert.Manager{
			Prompt: autocert.AcceptTOS,
			Cache:  autocert.DirCache("certs"),
		}

		server := &http.Server{
			Addr: listenAddr,
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
			},
			Handler: r,
		}
		go http.ListenAndServe(":http", certManager.HTTPHandler(nil))
		log.Fatal(server.ListenAndServeTLS("", ""))
	} else if selfSSL {

		server := &http.Server{
			Addr: listenAddr,
			TLSConfig: &tls.Config{
				GetCertificate: GetSelfSignedCertificate,
			},
			Handler: r,
		}
		log.Fatal(server.ListenAndServeTLS("", ""))
	} else {
		_ = browser.OpenURL(fmt.Sprintf("http://%s/", listenAddr))
		log.Println("Reverse Proxy Server running at " + listenAddr + "\n")
		log.Fatal(http.ListenAndServe(listenAddr, r))
	}
}
