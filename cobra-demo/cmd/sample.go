package cmd

type Config struct {
	ApiVersion string
	Kind       string
}

var Conf = new(Config)
