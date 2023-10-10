module github.com/dosemedia/pretest-api

// For local dev with crew-go in sibling folder:
// replace github.com/aaronblondeau/crew-go => ../crew-go

go 1.19

require (
	github.com/aaronblondeau/crew-go v1.0.5
	github.com/aws/aws-sdk-go-v2 v1.21.0
	github.com/aws/aws-sdk-go-v2/config v1.18.40
	github.com/aws/aws-sdk-go-v2/credentials v1.13.38
	github.com/aws/aws-sdk-go-v2/service/s3 v1.38.5
	github.com/disintegration/imaging v1.6.2
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.3.0
	github.com/iancoleman/strcase v0.0.0-20190422225806-e506e3ef7365
	github.com/joho/godotenv v1.5.1
	github.com/jordan-wright/email v4.0.1-0.20210109023952-943e75fe5223+incompatible
	github.com/k3a/html2text v1.2.1
	github.com/labstack/echo/v4 v4.11.1
	github.com/redis/go-redis/v9 v9.1.0
	github.com/shopspring/decimal v1.3.1
	github.com/steebchen/prisma-client-go v0.21.0
	github.com/takuoki/gocase v1.0.0
	golang.org/x/crypto v0.11.0
	golang.org/x/text v0.13.0
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.13 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.11 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.41 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.42 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.1.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.36 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.35 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.15.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.14.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.16.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.22.0 // indirect
	github.com/aws/smithy-go v1.14.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-co-op/gocron v1.27.1 // indirect
	github.com/go-redsync/redsync/v4 v4.8.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/labstack/gommon v0.4.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/image v0.12.0 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/sync v0.2.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/time v0.3.0 // indirect
)
