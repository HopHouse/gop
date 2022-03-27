package gopwebtampering

type OptionsStruct struct {
	Proxy string
}

var Options OptionsStruct

func init() {
	Options = OptionsStruct{}
}
