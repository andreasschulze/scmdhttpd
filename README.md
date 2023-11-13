# scmdhttpd

[![Actions Status](https://github.com/andreasschulze/scmdhttpd/workflows/Go%20Build/badge.svg)](https://github.com/andreasschulze/scmdhttpd/actions?query=workflow%3AGo%20Build)
[![Actions Status](https://github.com/andreasschulze/scmdhttpd/workflows/CodeQL/badge.svg)](https://github.com/andreasschulze/scmdhttpd/actions?query=workflow%3ACodeQL)
[![shellcheck](https://github.com/andreasschulze/scmdhttpd/actions/workflows/shellcheck.yml/badge.svg)](https://github.com/andreasschulze/scmdhttpd/actions/workflows/shellcheck.yml)
[![markdownlint](https://github.com/andreasschulze/scmdhttpd/actions/workflows/markdownlint.yml/badge.svg)](https://github.com/andreasschulze/scmdhttpd/actions/workflows/markdownlint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/andreasschulze/scmdhttpd)](https://goreportcard.com/report/github.com/andreasschulze/scmdhttpd)

Same/Simple Content for Multiple/Many Domains

Der Webserver ist für das Hosting von vielen Domains konzipiert. Dabei soll
überall der gleiche Inhalt ausgeliefert werden. Diese Inhalte bestehen aus
einer Eingangsseite (`index.html`), einem Favoriten-Icon (`favicon.ico`),
CSS-Formatierungen (`style.css`) und einer `robots.txt`, die von Suchmaschinen
ausgewertet wird.

Der Webserver antwortet nur für Servernamen, die in der Datei
`/data/vhosts.conf` gefunden werden. So wird verhindert, dass beliebige, dritte
Domains einen DNS-Namen mit der IP-Adresse des Servers propagieren. Es wird
HTTP und HTTPS unterstützt. Zertifikate werden von der CA Let's Encrypt
dynamisch bezogen und aktualisiert.

Neben dem Ausliefern von statischen Inhalten unterstützt der Webserver
HTTP-Redirects.  Der Modus wird pro Eintrag in der Datei `vhosts.conf` aktiviert,
wenn dort das Ziel als 2. Wert hinter einem Hostnamen angegeben wird.

## Optionen

- `--certificate_dir=<path>`

  Verzeichnis, in dem der Server TLS-Zertifikate dauerhaft speichern kann.

- `--staging`

  Der Server benutzt die 'staging' Umgebung von Let's Encrypt.

- `--acmeEndpoint=<acme-directory-url>`

  Statt Let's Encrypt kann hier [eine andere ACME-Instanz](https://datatracker.ietf.org/doc/html/rfc8555#section-7.1.1)
  konfiguriert werden. In diesem Fall ist die Option `--staging` belanglos.

- `--datadir=<path>`

  Die Dateien `vhosts.conf`, `index.html`, `favicon.ico`, `style.css` und
  `robots.txt` werden im Verzeichnis `/data` gesucht, wenn nicht mit `--datadir`
  ein alternativer Pfad angegeben wird.

## Dateien in /data

- `vhosts.conf`

  Liste mit Hostnamen, für die der Webserver Inhalte liefert. Format: ein
  Name pro Zeile, kein abschließender Punkt. Optional kann als 2. Wert ein
  Weiterleitungsziel angegeben werden.

  Beispiel:

  ```txt
  example
  example.org https://example.net/foo
  ```

  Mit diesen Einträgen beantwortet der Service Anfragen für

  - `http://www.example` mit einem Redirect nach `https://www.example`
  - `https://www.example` mit einem Redirect nach `https://example`
  - `http://example` mit einem Redirect nach `https://example`
  - `https://example` mit Inhalten (index.html, ...)
  - `http://www.example.org` mit einem Redirect nach `https://www.example.org`
  - `https://www.example.org` mit einem Redirect nach `https://example.net/foo`
  - `http://example.org` mit einem Redirect nach `https://example.org`
  - `https://example.org` mit einem Redirect nach `https://example.net/foo`

  Wird die Datei geändert, muss der Server neu gestartet werden. Hostnamen
  (Spalte 1) werden beim Start des Servers in Kleinbuchstaben konvertiert.

- `index.html`

  HTML-Seite, die beim Aufruf der URL `/` (und `/index.html`) ausgegeben wird.

- `robots.txt`

  Text-Datei, die beim Aufruf der URL `/robots.txt` ausgegeben wird.

- `favicon.ico`

  Icon-Datei, die beim Aufruf der URL `/favicon.ico` ausgegeben wird.

- `style.css`

  CSS-Datei, die beim Aufruf der URL `/style.css` ausgegeben wird.

Werden die genannten URLs per HTTP aufgerufen, erfolgt ein
[permanenter Redirect](https://datatracker.ietf.org/doc/html/rfc7231#section-6.4.2)
auf die entsprechende HTTPS-URL.

Alle anderen URLs werden mit [404 Not Found](https://datatracker.ietf.org/doc/html/rfc7231#section-6.5.4)
beantwortet.

## Anfragen für unbekannte Hostnamen

HTTP-Anfragen an den Server mit einem Hostnamen, der nicht in `vhosts.conf`
konfiguriert ist, werden mit [400 Bad Request](https://datatracker.ietf.org/doc/html/rfc7231#section-6.5.1)
beantwortet; bei HTTPS-Anfragen kommt keine TLS-Verbindung zustande.

## Logging

Anfragen werden auf STDOUT geloggt. Das Format entspricht weitgehend dem
[`combined` Logformat eines NGINX Webservers](https://nginx.org/r/log_format).

Ausnahmen:

- an 3. Stelle wird anstatt [`$remote_user`](https://nginx.org/en/docs/http/ngx_http_core_module.html#var_remote_user)
  der Hostname [`$host`](https://nginx.org/en/docs/http/ngx_http_core_module.html#var_host)
  geloggt.

- an 7. Stelle wird die Anzahl der Antwortbytes ([`$body_bytes_sent`](https://nginx.org/en/docs/http/ngx_http_core_module.html#var_body_bytes_sent))
  immer mit 42 geloggt.

- bei HTTPS-Anfragen werden nach dem User-Agent TLS-Version sowie TLS-Cipher
  im Format "tlsversion=... tlscipher=..." ausgegeben.
