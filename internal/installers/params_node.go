package installers

import "errors"

type NodeParams struct {
	Endpoint string
	NodeId   string
	Secret   string
}

func (this *NodeParams) Validate() error {
	if len(this.Endpoint) == 0 {
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
