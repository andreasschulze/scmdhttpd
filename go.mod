module github.com/andreasschulze/scmdhttpd

go 1.18

require golang.org/x/crypto v0.0.0-00010101000000-000000000000

require (
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 // indirect
	golang.org/x/text v0.3.6 // indirect
)

replace golang.org/x/crypto => github.com/hochhaus/crypto v0.0.0-20220107215148-ba7d272dee0b
