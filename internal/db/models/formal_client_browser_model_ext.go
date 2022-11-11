package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
)

// DecodeCodes 解析代号
func (this *FormalClientBrowser) DecodeCodes() []string {
	if IsNull(this.Codes) {
		return nil
	}

	var result = []string{}
	err := json.Unmarshal(this.Codes, &result)
	if err != nil {
		remotelogs.Error("FormalClientBrowser.DecodeCodes", err.Error())
	}

	return result
}
