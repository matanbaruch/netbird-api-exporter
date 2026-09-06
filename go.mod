module github.com/matanbaruch/netbird-api-exporter

go 1.27.0

require (
	github.com/netbirdio/netbird v0.78.0
	github.com/prometheus/client_golang v1.24.1
	github.com/sirupsen/logrus v1.10.2
)

require (
	github.com/DeRuina/timberjack v1.4.2 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.19.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oapi-codegen/runtime v1.1.2 // indirect
	github.com/petermattis/goid v0.0.0-20250303134427-723919f7f203 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.70.1 // indirect
	github.com/prometheus/procfs v0.21.1 // indirect
	github.com/skratchdot/open-golang v0.0.0-20200116055534-eef842397966 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.zx2c4.com/wireguard/windows v0.5.3 // indirect
	google.golang.org/grpc v1.80.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

// Mirrors netbird's own replace directives: they are not inherited by dependents, so `go mod tidy` cannot resolve dex/server/signer. Keep in sync with netbird's go.mod.
replace github.com/dexidp/dex => github.com/netbirdio/dex v0.244.1-0.20260512110716-8d70ad8647c1

replace github.com/dexidp/dex/api/v2 => github.com/netbirdio/dex/api/v2 v2.0.0-20260512110716-8d70ad8647c1
