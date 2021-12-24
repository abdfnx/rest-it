package options

type Auth struct {
	TokenAuth 		  string
	BasicAuthUsername string
	BasicAuthPassword string
	Type 		      string
}

type Method struct {
	AuthType 	    *Auth
	JustShowBody    bool
	JustShowHeaders bool
	SaveFile 		string
	ContentType 	string
	OpenEditor 		bool
	Body 			string
	IsBodyStdin 	bool
}

type CLIOptions struct {
	Method *Method
	URL    string
}

type InstallCommandOptions struct {
	Shell string
	URL   string
}

type RunCommandOptions struct {
	Path    string
	ShowAll bool
}

type GetLatestCommandOptions struct {
	Registry  string
	Repo      string
	Token     string
}
