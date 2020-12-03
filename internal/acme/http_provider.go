package acme

type HTTPProvider struct {
	onAuth AuthCallback
}

func NewHTTPProvider(onAuth AuthCallback) *HTTPProvider {
	return &HTTPProvider{
		onAuth: onAuth,
	}
}

func (this *HTTPProvider) Present(domain, token, keyAuth string) error {
	if this.onAuth != nil {
		this.onAuth(domain, token, keyAuth)
	}
	//http01.ChallengePath()
	return nil
}

func (this *HTTPProvider) CleanUp(domain, token, keyAuth string) error {
	return nil
}
