
// https://github.com/danmarg/sts-mate
// https://blog.cloudflare.com/exposing-go-on-the-internet/
// https://bruinsslot.jp/post/go-secure-webserver/
// https://blog.kowalczyk.info/article/Jl3G/https-for-free-in-go-with-little-help-of-lets-encrypt.html

package main

import (
        "golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"

	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"crypto/tls"
)

const (
	certsDir = "certs"
)

var (
	certdir           = flag.String("certificate_dir", "certificate-dir", "Directory in which to store certificates.")
	acmeEndpoint      = flag.String("acme_endpoint", "", "If set, uses a custom ACME endpoint URL. It doesn't make sense to use this with --staging.")
	staging	          = flag.Bool("staging", false, "If true, uses Let's Encrypt 'staging' environment instead of prod.")
	datadir           = flag.String("data_dir", "/data", "Directory where vhosts.conf, index.html, robots.txt an favicon.ico are found")

	// global var
	vhosts            []string

)

func hostPolicy() autocert.HostPolicy {
	return func(ctx context.Context, host string) error {
		if !contains(vhosts, strings.ToLower(host)) {
			return fmt.Errorf("host %s not listed in %s/vhosts.conf", host, *datadir)
		}
		return nil
	}
}

// https://play.golang.org/p/Qg_uv_inCek
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func log(r *http.Request, response_status_code int) {
	var ts = time.Now().Format("02/Jan/2006:15:04:05 -0700")
	var clientIP string = r.RemoteAddr
	if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
		clientIP = clientIP[:colon]
	}
	clientIP = strings.ReplaceAll(strings.ReplaceAll(clientIP, "[", ""), "]", "")
	fmt.Printf("%s - %s [%s] \"%s %s %s\" %d 42 \"%s\" \"%s\"\n", clientIP, r.Host, ts, r.Method, r.RequestURI, r.Proto, response_status_code, r.Header.Get("Referer"), r.UserAgent())
}

func main() {

	flag.Parse()

	content, err := ioutil.ReadFile(*datadir + "/vhosts.conf")
	if err != nil {
		panic(err)
	}
	// aus dem Dateiinhalt eine einen String aus Kleinbustaben machen
	vhosts_file := strings.ToLower(string(content))

	// Slice/Array von einzelnen Hostnamen
	vhosts = strings.Split(string(vhosts_file), "\n")

	if *staging {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		var lHost = strings.ToLower(r.Host)
		var lDomain = lHost
		var redir2domain bool = false
		var scheme string = "http"
		if strings.HasPrefix(lHost, "www.") {
			lDomain = strings.Replace(lHost, "www.", "", 1)
			redir2domain = true
		}
		if r.TLS != nil {
			scheme += "s"
		}
		if !contains(vhosts, lDomain) {
			log(r, http.StatusBadRequest)
			http.Error(w, "400 bad request", http.StatusBadRequest)
		} else {
			switch r.URL.Path {
			case "/index.html": fallthrough
			case "/favicon.ico": fallthrough
			case "/robots.txt": fallthrough
                        case "/style.css": fallthrough
			case "/":
				if r.TLS == nil {
					w.Header().Set("Connection", "close")
					http.Redirect(w, r, "https://" + lHost + r.URL.Path, http.StatusMovedPermanently)
					log(r, 301)
					return
				}
				if redir2domain {
					w.Header().Set("Connection", "close")
					http.Redirect(w, r, scheme + "://" + lDomain + r.URL.Path, http.StatusMovedPermanently)
					log(r, 301)
					return
				}
				w.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubdomains")
				w.Header().Add("Content-Security-Policy", "default-src 'self'")
				w.Header().Add("X-XSS-Protection", "1; mode=block")
				w.Header().Add("X-Frame-Options", "DENY")
				w.Header().Add("Referrer-Policy", "strict-origin-when-cross-origin")
				w.Header().Add("X-Content-Type-Options", "nosniff")
				w.Header().Add("Expect-CT", "max-age=6048000,enforce")

				w.Header().Add("Permissions-Policy", "interest-cohort=()")

				w.Header().Add("Cache-Control", "public; max-age=86400")

				if r.URL.Path == "/" {
					http.ServeFile(w, r, *datadir + "/index.html")
				} else {
					http.ServeFile(w, r, *datadir + r.URL.Path)
				}
				log(r, 200)
				return
			}
			http.NotFound(w, r)
		}
	})

	cm := &autocert.Manager {
		Cache:		autocert.DirCache(filepath.Join(*certdir, certsDir)),
		Prompt:		autocert.AcceptTOS,
		HostPolicy:	hostPolicy(),
	}

	if *acmeEndpoint != "" {
		cm.Client = &acme.Client{DirectoryURL: *acmeEndpoint}
	} else if *staging {
		cm.Client = &acme.Client{DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory"}
	}

        srv_plain := &http.Server {
		Addr: ":http",
		Handler: http.DefaultServeMux,
	}
	srv_tls := &http.Server {
		Addr: ":https",
                Handler: http.DefaultServeMux,
		TLSConfig: cm.TLSConfig(),
        }

	srv_tls.TLSConfig.MinVersion = tls.VersionTLS12
	srv_tls.TLSConfig.PreferServerCipherSuites = true
	srv_tls.TLSConfig.CipherSuites = []uint16 {
		// same as 'openssl11 cipher -v "ECDHE+AESGCM:ECDHE+CHACHA20"' without RSA
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	}

	// serve http
	go func() {
		fmt.Fprintln(os.Stderr, srv_plain.ListenAndServe())
	}()

	// serve https
	fmt.Fprintln(os.Stderr, srv_tls.ListenAndServeTLS("", ""))

	os.Exit(1)
}
