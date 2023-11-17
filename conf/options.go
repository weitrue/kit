package conf

// Option defines the method to customize the config options.
type Option func(opt *options)

type options struct {
	env bool
}

// UseEnv customizes the config to use environment variables.
func UseEnv() Option {
	return func(opt *options) {
		opt.env = true
	}
}
