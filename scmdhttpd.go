// https://github.com/danmarg/sts-mate
// https://blog.cloudflare.com/exposing-go-on-the-internet/
// https://bruinsslot.jp/post/go-secure-webserver/
// https://blog.kowalczyk.info/article/Jl3G/https-for-free-in-go-with-little-help-of-lets-encrypt.html

package main

import (
	"encoding/csv"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"

	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	certsDir       = "certs"
	programName    = "scmdHTTPd"
	programVersion = "2.3.2"
)

var (
	certdir      = flag.String("certificate_dir", "certificate-dir", "Directory in which to store certificates.")
	acmeEndpoint = flag.String("acme_endpoint", "", "If set, uses a custom ACME endpoint URL. It doesn't make sense to use this with --staging.")
	staging      = flag.Bool("staging", false, "If true, uses Let's Encrypt 'staging' environment instead of prod.")
	datadir      = flag.String("data_dir", "/data", "Directory where vhosts.conf, index.html, robots.txt, security.txt and favicon.ico are found")
	version      = flag.Bool("version", false, "print version and exit.")

	// global var
	vhosts = make(map[string]string)
)

func hostPolicy() autocert.HostPolicy {
	return func(ctx context.Context, host string) error {
		// $domain und www.$domain ist ok
		host = strings.Replace(strings.ToLower(host), "www.", "", 1)
		_, knownVhost := vhosts[host]
		if !knownVhost {
			return fmt.Errorf("host %s not listed in %s/vhosts.conf", host, *datadir)
		}
		return nil
	}
}

func log(r *http.Request, responseStatusCode int) {
	var ts = time.Now().Format("02/Jan/2006:15:04:05 -0700")
	var clientIP string = r.RemoteAddr
	if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
		clientIP = clientIP[:colon]
	}
	clientIP = strings.ReplaceAll(strings.ReplaceAll(clientIP, "[", ""), "]", "")
	// https://owasp.org/www-community/attacks/Log_Injection
	userAgent := strings.Replace(r.UserAgent(), "\n", "", -1)
	userAgent = strings.Replace(userAgent, "\r", "", -1)
	referer := strings.Replace(r.Header.Get("Referer"), "\n", "", -1)
	referer = strings.Replace(referer, "\r", "", -1)
	if r.TLS == nil {
		fmt.Printf("%s - %s [%s] \"%s %s %s\" %d 42 \"%s\" \"%s\"\n", clientIP, r.Host, ts, r.Method, r.RequestURI, r.Proto, responseStatusCode, referer, userAgent)
	} else {
		// https://pkg.go.dev/crypto/tls#ConnectionState
		// https://gist.github.com/xjdrew/97be3811966c8300b724deabc10e38e2
		// https://github.com/golang/go/issues/46308
		var VersionName string
		switch r.TLS.Version {
		// names like postfix use, too
		case 0x303:
			VersionName = "TLSv1.2"
		case 0x304:
			VersionName = "TLSv1.3"
		default:
			VersionName = "?"
		}
		fmt.Printf("%s - %s [%s] \"%s %s %s\" %d 42 \"%s\" \"%s\" tlsversion=%s tlscipher=%s\n", clientIP, r.Host, ts, r.Method, r.RequestURI, r.Proto, responseStatusCode, referer, userAgent, VersionName, tls.CipherSuiteName(r.TLS.CipherSuite))
	}
}

func readcsvfile(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ' '
	r.Comment = '#'
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true

	entries, err := r.ReadAll()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		key := strings.ToLower(entry[0])
		if len(entry) == 1 {
			vhosts[key] = ""
		} else if len(entry) >= 2 {
			vhosts[key] = entry[1]
		} else {
			fmt.Printf("ignore %v\n", entry)
		}
	}

	return nil // keinFehler
}

func versionInfo(prefix string) {
	// prefix is empty or "starting "
	fmt.Printf("%s%s-%s, built with %s\n", prefix, programName, programVersion, runtime.Version())
}

func main() {

	flag.Parse()

	if *version {
		versionInfo("")
		os.Exit(0)
	}

	err := readcsvfile(*datadir + "/vhosts.conf")
	if err != nil {
		panic(err)
	}

	if *staging {
		/*
		 * CodeQL bemängelt, dass hier Zertifikatsprüfung ausgeschaltet wird
		 * wenn --staging benutzt wird, muss das aber sein
		 * Daher wird die CodeQL-Warning hier deaktiviert
		 */
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // CodeQL [go/disabled-certificate-check]
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
		// https://blog.golang.org/maps: A two-value assignment tests for the existence of a key
		redir301, knownDomain := vhosts[lDomain]
		if !knownDomain {
			log(r, http.StatusBadRequest)
			fmt.Printf("host %s not listed in %s/vhosts.conf\n", lDomain, *datadir)
			http.Error(w, "400 bad request", http.StatusBadRequest)
		} else if redir301 != "" {
			if r.TLS == nil {
				w.Header().Set("Connection", "close")
				http.Redirect(w, r, "https://"+lHost+"/", http.StatusMovedPermanently)
				log(r, 301)
				return
			}
			w.Header().Set("Connection", "close")
			http.Redirect(w, r, redir301, http.StatusMovedPermanently)
			log(r, 301)
			return
		} else {
			switch r.URL.Path {
			case "/index.html":
				fallthrough
			case "/favicon.ico":
				fallthrough
			case "/robots.txt":
				fallthrough
			case "/style.css":
				fallthrough
			case "/.well-known/security.txt":
				fallthrough
			case "/":
				if r.TLS == nil {
					w.Header().Set("Connection", "close")
					http.Redirect(w, r, "https://"+lHost+r.URL.Path, http.StatusMovedPermanently)
					log(r, 301)
					return
				}
				if redir2domain {
					w.Header().Set("Connection", "close")
					http.Redirect(w, r, scheme+"://"+lDomain+r.URL.Path, http.StatusMovedPermanently)
					log(r, 301)
					return
				}
				w.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubdomains")
				w.Header().Add("Content-Security-Policy", "default-src 'none';img-src 'self';style-src 'self';form-action 'none';base-uri 'self';frame-ancestors 'none';upgrade-insecure-requests;")
				w.Header().Add("X-Xss-Protection", "0")
				w.Header().Add("X-Frame-Options", "DENY")
				w.Header().Add("Referrer-Policy", "no-referrer")
				w.Header().Add("X-Content-Type-Options", "nosniff")
				w.Header().Add("Expect-Ct", "max-age=6048000,enforce")

				w.Header().Add("Permissions-Policy", "interest-cohort=()")

				w.Header().Add("Cache-Control", "max-age=86400, must-revalidate")

				if r.URL.Path == "/" {
					http.ServeFile(w, r, *datadir+"/index.html")
				} else if r.URL.Path == "/.well-known/security.txt" {
					http.ServeFile(w, r, *datadir+"/security.txt")
				} else {
					http.ServeFile(w, r, *datadir+r.URL.Path)
				}
				log(r, 200)
				return
			}
			log(r, 404)
			http.NotFound(w, r)
		}
	})

	cm := &autocert.Manager{
		Cache:      autocert.DirCache(filepath.Join(*certdir, certsDir)),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: hostPolicy(),
	}

	if *acmeEndpoint != "" {
		cm.Client = &acme.Client{DirectoryURL: *acmeEndpoint}
	} else if *staging {
		cm.Client = &acme.Client{DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory"}
	}

	srvPlain := &http.Server{
		Addr:    ":http",
		Handler: http.DefaultServeMux,
	}
	srvTLS := &http.Server{
		Addr:      ":https",
		Handler:   http.DefaultServeMux,
		TLSConfig: cm.TLSConfig(),
	}

	srvTLS.TLSConfig.MinVersion = tls.VersionTLS12
	srvTLS.TLSConfig.PreferServerCipherSuites = true
	srvTLS.TLSConfig.CipherSuites = []uint16{
		// same as 'openssl11 cipher -v "ECDHE+AESGCM:ECDHE+CHACHA20"' without RSA
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	}
	// disable http2 and higher
	srvTLS.TLSConfig.NextProtos = []string{"http/1.0", "http/1.1", "acme-tls/1"}

	versionInfo("starting ")

	// serve http
	go func() {
		fmt.Fprintln(os.Stderr, srvPlain.ListenAndServe())
	}()

	// serve https
	fmt.Fprintln(os.Stderr, srvTLS.ListenAndServeTLS("", ""))

	os.Exit(1)
}
