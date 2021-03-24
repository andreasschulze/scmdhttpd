# scmdhttpd - Same/Simple Content for Multilple/Many Domains

Der Webserver ist für das Hosting von vielen Domains konzipiert. Dabei soll
überall der gleiche Inhalt ausgeliefert werden. Diese Inhalte bestehen aus
einer Eingangsseite (`index.html`), einem Favoriten-Icon (`favicon.ico`) und
einer `robots.txt`, die von Suchmaschinen ausgewertet wird.

Der Webserver antwortet nur für Servernamen, die in der Datei
`/data/vhosts.conf` gefunden werden. So wird verhindert, dass beliebige, dritte
Domains einen DNS-Namen mit der IP-Adresse des Servers propagieren. Es wird
HTTP und HTTPS unterstützt. Zertifikate werden von der CA Let's Encrypt
dynamisch bezogen und aktualisiert.

## Optionen

* `--certificate_dir=<path>`

  Verzeichnis, in dem der Server TLS-Zertifikate dauerhaft speichern kann.

* `--staging`

  Der Server benutzt die 'staging' Umgebung von Let's Encrypt.

* `--acmeEndpoint=<acme-directory-url>`

  Statt Let's Encrypt kann hier [eine andere ACME-Instanz](https://tools.ietf.org/html/rfc8555#section-7.1.1)
  konfiguriert werden. In diesem Fall ist die Option `--staging` belanglos.

* `datadir=<path>

  Die Dateien `vhosts.conf`, `index.html`, `favicon.ico` und `robots.txt`
  werden im Verzeichnis `/data` gesucht, wenn nicht mit `--datadir` ein
  alternativer Pfad angegeben wird.

## Dateien in /data

* `vhosts.conf`

  Liste mit Hostnamen, für die der Webserver Inhalte liefert. Format: ein
  Name pro Zeile, kein abschließender Punkt. Beispiel:

  ```txt
  example
  www.example
  ```

  Wird die Datei geändert, muss der Server neu gestartet werden.

* `index.html`

  HTML-Seite, die beim Aufruf der URL `/` ausgegeben wird.

* `robots.txt`

  Text-Datei, die beim Aufruf der URL `/robots.txt` ausgegeben wird.

* `favicon.ico`

  Icon-Datei, die beim Aufruf der URL `/favicon.ico` ausgegeben wird.

Werden die genannten URLs per HTTP aufgerufen, erfolgt ein
[permanenter Redirect](https://tools.ietf.org/html/rfc7231#section-6.4.2)
auf die entsprechnde HTTPS-URL.

Alle anderen URLs werdem mit [404 Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)
beantwortet.
