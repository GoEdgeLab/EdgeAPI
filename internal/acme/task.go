package acme

import "github.com/TeaOSLab/EdgeAPI/internal/dnsclients"

type AuthType = string

const (
	AuthTypeDNS  AuthType = "dns"
	AuthTypeHTTP AuthType = "http"
)

type Task struct {
	Provider *Provider
	Account  *Account
	User     *User
	AuthType AuthType
	Domains  []string

	// DNS相关
	DNSProvider dnsclients.ProviderInterface
	DNSDomain   string
}
