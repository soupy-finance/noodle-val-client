module github.com/soupy-finance/noodle-val-client

go 1.17

require (
	github.com/soupy-finance/noodle v0.0.1 
	github.com/ignite/cli v0.22.1
)

replace github.com/soupy-finance/noodle => ../noodle
replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
