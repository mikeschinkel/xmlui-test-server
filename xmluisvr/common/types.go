package common

type (
	DirPath         string //Absolute or Relative
	EnvironmentVars map[string]string
	ExtensionId     string

	Filepath      string //Absolute or Relative
	ConnectString string //Absolute or Relative
	QueryString   string //Absolute or Relative
	FullURL       string
	Identifier    string //Absolute or Relative
	JSONBytes     []byte //Absolute or Relative

	LoadOrder  int
	OSArch     string
	SHA256     string
	Version    string
	Host       string
	ServerPort int
)

type ContentGetter interface {
	Content() any
}
