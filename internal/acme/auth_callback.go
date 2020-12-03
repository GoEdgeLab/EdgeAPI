package acme

type AuthCallback func(domain, token, keyAuth string)
