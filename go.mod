module github.com/hujun-open/myflags

go 1.20

require (
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/exp v0.0.0-20230801115018-d63ba01acd4b
)

require github.com/inconshreveable/mousetrap v1.1.0 // indirect

replace github.com/spf13/pflag => ../pflag
