package installers

type Credentials struct {
	Host       string
	Port       int
	Username   string
	Password   string
	PrivateKey string
	Passphrase string
	Method     string
	Sudo       bool
}
