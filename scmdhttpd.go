
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
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
//	"time"
	"crypto/tls"
)

const (
	certsDir = "certs"
//	hostsDir = "hosts"
)

var (
	certdir           = flag.String("certificate_dir", "certificate-dir", "Directory in which to store certificates.")
	acmeEndpoint      = flag.String("acme_endpoint", "", "If set, uses a custom ACME endpoint URL. It doesn't make sense to use this with --staging.")
	staging	          = flag.Bool("staging", false, "If true, uses Let's Encrypt 'staging' environment instead of prod.")
//	tryCertNoMoreThan = flag.Duration("try_cert_no_more_often_than", 24*time.Hour, "Don't try to request a cert for a host more often than this.")
	datadir           = flag.String("data_dir", "/data", "Directory where vhosts.conf, index.html, robots.txt an favicon.ico are found")

	// global var
	vhosts            string

)

func hostPolicy() autocert.HostPolicy {
	return func(ctx context.Context, host string) error {
		if !strings.Contains(vhosts, strings.ToLower(host)) {
			return fmt.Errorf("host %s not listed in %s/vhosts.conf", host, datadir)
		}
		return nil
/*		hdir := filepath.Join(*certdir, hostsDir)
		p := filepath.Join(hdir, filepath.Clean(host))
		if s, err := os.Stat(p); err != nil {
			if !os.IsNotExist(err) {
				// Some other unexpected error here.
				return err
			}
		} else if s != nil && time.Now().Sub(s.ModTime()) < *tryCertNoMoreThan {
			// Too recently attempted this host.
			return fmt.Errorf("too recently attempted host %s", host)
		}
		// Touch the host file.
		if _, err := os.Stat(hdir); os.IsNotExist(err) {
			if err := os.MkdirAll(hdir, 0700); err != nil {
				return err
			}
		}
		_, err := os.Create(p)
		return err */
	}
}

func main() {

	flag.Parse()

	vhosts_file, err := ioutil.ReadFile(*datadir + "/vhosts.conf")
	if err != nil {
		panic(err)
	}
	vhosts = string(vhosts_file)

	if *staging {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		if strings.Contains(vhosts, strings.ToLower(r.Host)) {
			switch r.URL.Path {
			case "/robots.txt": fallthrough
			case "/favicon.ico": fallthrough
			case "/":
				var scheme string = "http"
				if r.TLS == nil {
					w.Header().Set("Connection", "close")
					http.Redirect(w, r, "https://" + r.Host + r.URL.Path, http.StatusMovedPermanently)
				} else {
					w.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubdomains")
					w.Header().Add("Content-Security-Policy", "default-src 'none'")
					w.Header().Add("X-XSS-Protection", "1; mode=block")
					w.Header().Add("X-Frame-Options", "DENY")
					w.Header().Add("Referrer-Policy", "strict-origin-when-cross-origin")
					w.Header().Add("X-Content-Type-Options", "nosniff")
					w.Header().Add("Expect-CT", "max-age=6048000,enforce")

					scheme += "s"
					if r.URL.Path == "/" {
						http.ServeFile(w, r, *datadir + "/index.html")
					} else {
						http.ServeFile(w, r, *datadir + r.URL.Path)
					}
				}
				log.Printf("%s : %s://%s%s : %s\n", r.RemoteAddr, scheme, r.Host, r.URL.Path, r.UserAgent())
				return
			}
		}
		http.NotFound(w, r)
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

	// server https
	fmt.Fprintln(os.Stderr, srv_tls.ListenAndServeTLS("", ""))

	os.Exit(1)
}
