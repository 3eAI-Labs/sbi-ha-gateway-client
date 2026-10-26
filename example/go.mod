module github.com/3eAI-Labs/sbi-ha-gateway-client/example

go 1.26.0

require (
	github.com/3eAI-Labs/sbi-ha-gateway-client v0.0.0
	github.com/nats-io/nats.go v1.48.0
)

require (
	github.com/klauspost/compress v1.18.7 // indirect
	github.com/nats-io/nkeys v0.4.12 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
)

// The client module is this repository's root.
replace github.com/3eAI-Labs/sbi-ha-gateway-client => ../
