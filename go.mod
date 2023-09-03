module aaronblondeau.com/hasura-base-go

// For local dev with crew-go in sibling folder:
replace github.com/aaronblondeau/crew-go => ../../Dose/crew-go

go 1.19

require (
	github.com/aaronblondeau/crew-go v1.0.4
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/iancoleman/strcase v0.0.0-20190422225806-e506e3ef7365
	github.com/joho/godotenv v1.5.1
	github.com/labstack/echo/v4 v4.11.1
	github.com/redis/go-redis/v9 v9.1.0
	github.com/shopspring/decimal v1.3.1
	github.com/steebchen/prisma-client-go v0.21.0
	github.com/takuoki/gocase v1.0.0
	golang.org/x/crypto v0.11.0
	golang.org/x/text v0.11.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-co-op/gocron v1.27.1 // indirect
	github.com/go-redsync/redsync/v4 v4.8.1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/labstack/gommon v0.4.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/sync v0.2.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/time v0.3.0 // indirect
)
