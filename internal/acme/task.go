package acme

import "github.com/TeaOSLab/EdgeAPI/internal/dnsclients"

type Task struct {
	User        *User
	DNSProvider dnsclients.ProviderInterface
	DNSDomain   string
	Domains     []string
}
