package gopproxy

type Options struct {
	Host string
	Port string

	VerboseOption bool
	InterceptOption bool
}

var (
	InterceptMode bool
	InterceptChan chan bool
)

func InitOptions(host string, port string, verboseOption bool, interceptOption bool) Options {
	var options Options

	options.Host = host
	options.Port = port
	options.VerboseOption = verboseOption
	options.InterceptOption = interceptOption

	return options
}