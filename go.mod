module github.com/f4ke-n0name/autoparts-hub

go 1.25.8

require (
	github.com/Masterminds/squirrel v1.5.4
	github.com/go-chi/chi/v5 v5.2.5
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/jackc/pgx/v5 v5.9.1
	github.com/joho/godotenv v1.5.1
	github.com/swaggo/http-swagger v1.3.4
	github.com/swaggo/swag v1.16.6
	github.com/twmb/franz-go v1.20.7
	golang.org/x/crypto v0.50.0
	golang.org/x/sync v0.20.0
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.6 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.18.4 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/pierrec/lz4/v4 v4.1.25 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/swaggo/files v0.0.0-20220610200504-28940afbdbfe // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.12.0 // indirect
	golang.org/x/mod v0.34.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	golang.org/x/tools v0.43.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260406210006-6f92a3bedf2d // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/f4ke-n0name/autoparts-hub/pkg/jwt => ../../pkg/jwt

replace github.com/f4ke-n0name/autoparts-hub/pkg/logger => ../../pkg/logger

replace github.com/f4ke-n0name/autoparts-hub/pkg/errors => ../../pkg/errors
