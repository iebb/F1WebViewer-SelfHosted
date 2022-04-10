package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pkg/browser"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

var builtInSSL = false
var selfSSL = false
var port = 13331

const RSABits = 2048
const ValidFor = time.Hour * 1440

var certCache sync.Map

func GetSelfSignedCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if cert, ok := certCache.Load(clientHello.ServerName); ok {
		return cert.(*tls.Certificate), nil
	}
	priv, err := rsa.GenerateKey(rand.Reader, RSABits)
	if err != nil {
		return nil, err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(ValidFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	host := clientHello.ServerName
	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, host)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	cert := &tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  priv,
	}
	certCache.Store(clientHello.ServerName, cert)
	return cert, err
}

func init() {
	flag.BoolVar(&builtInSSL, "ssl", false,
		"built-in ssl support. this requires a public IP, and port 80 being available in order to fetch the certs",
	)
	flag.BoolVar(&selfSSL, "self-sign", false,
		"self-signed ssl support. would display a warning on browsers",
	)
	flag.IntVar(&port, "port", 13331, "port number")
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
	r.HandleFunc("/proxy/{url:https?://?.*}", handler)
	r.PathPrefix("/authenticate").HandlerFunc(fnHandler)
	r.PathPrefix("/66571939").HandlerFunc(fnHandler)
	r.PathPrefix("/").HandlerFunc(staticHandler)
	r.SkipClean(true)

	listenAddr := fmt.Sprintf(":%d", port)

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
		_ = browser.OpenURL(fmt.Sprintf("http://localhost:%d/", port))
		log.Println("Reverse Proxy Server running at " + listenAddr + "\n")
		log.Fatal(http.ListenAndServe(listenAddr, r))
	}
}
