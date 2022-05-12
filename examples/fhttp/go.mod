module github.com/sleeyax/gotcha/examples/fhttp

go 1.16

require (
	github.com/sleeyax/gotcha v0.1.1
	github.com/sleeyax/gotcha/adapters/fhttp v0.0.0-00010101000000-000000000000
	github.com/useflyent/fhttp v0.0.0-20211004035111-333f430cfbbf
)

replace (
	github.com/sleeyax/gotcha => ../..
	github.com/sleeyax/gotcha/adapters/fhttp => ../../adapters/fhttp
)
