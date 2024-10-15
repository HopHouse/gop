package gopproxy

type Options struct {
	Host string
	Port string

	caFileOption        string
	caPrivKeyFileOption string

	VerboseOption   bool
	InterceptOption bool
}

var (
	InterceptMode bool
	InterceptChan chan bool
)

func InitOptions(host string, port string, verboseOption bool, interceptOption bool, caFileOption string, caPrivKeyFileOption string) Options {
	var options Options

	options.Host = host
	options.Port = port
	options.caFileOption = caFileOption
	options.caPrivKeyFileOption = caPrivKeyFileOption
	options.VerboseOption = verboseOption
	options.InterceptOption = interceptOption

	return options
}
