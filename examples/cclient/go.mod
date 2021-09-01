module github.com/sleeyax/gotcha/examples/cclient

go 1.16

require (
	github.com/refraction-networking/utls v0.0.0-20210713165636-0b2885c8c0d4
	github.com/sleeyax/gotcha v0.0.2
	github.com/sleeyax/gotcha/adapters/cclient v0.0.0-20210627011908-60a6c193f1b8
)

replace (
	github.com/sleeyax/gotcha => ../..
	github.com/sleeyax/gotcha/adapters/cclient => ../../adapters/cclient
)
