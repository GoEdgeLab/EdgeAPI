package installers

import (
	"errors"
	"strings"
)

type NodeParams struct {
	Endpoints []string
	NodeId    string
	Secret    string
}

func (this *NodeParams) Validate() error {
	if len(this.Endpoints) == 0 {
		return errors.New("'endpoint' should not be empty")
	}
	if len(this.NodeId) == 0 {
		return errors.New("'nodeId' should not be empty")
	}
	if len(this.Secret) == 0 {
		return errors.New("'secret' should not be empty")
	}
	return nil
}

func (this *NodeParams) QuoteEndpoints() string {
	if len(this.Endpoints) == 0 {
		return ""
	}
	return "\"" + strings.Join(this.Endpoints, "\", \"") + "\""
}
