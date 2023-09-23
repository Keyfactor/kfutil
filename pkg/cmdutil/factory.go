package cmdutil

type PersistentFlags struct {
	ConfigFile   string
	NoPrompt     bool
	Experimental bool
	Debug        bool
	Profile      string
	Username     string
	Hostname     string
	Password     string
	Domain       string
	APIPath      string
}

type Factory interface {
	PersistentFlags
}
