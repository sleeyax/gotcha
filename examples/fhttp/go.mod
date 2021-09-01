module github.com/sleeyax/gotcha/examples/fhttp

go 1.16

require (
	github.com/sleeyax/gotcha v0.0.2
	github.com/sleeyax/gotcha/adapters/fhttp v0.0.0-00010101000000-000000000000
	github.com/useflyent/fhttp v0.0.0-20210801005649-f160dd923789
)

replace (
	github.com/sleeyax/gotcha => ../..
	github.com/sleeyax/gotcha/adapters/fhttp => ../../adapters/fhttp
)
