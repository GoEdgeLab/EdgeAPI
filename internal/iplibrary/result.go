package iplibrary

import (
	"github.com/iwind/TeaGo/lists"
	"strings"
)

type Result struct {
	CityId   int64
	Country  string
	Region   string
	Province string
	City     string
	ISP      string
}

func (this *Result) Summary() string {
	pieces := []string{}
	if len(this.Country) > 0 {
		pieces = append(pieces, this.Country)
	}
	if len(this.Province) > 0 && !lists.ContainsString(pieces, this.Province) {
		pieces = append(pieces, this.Province)
	}
	if len(this.City) > 0 && !lists.ContainsString(pieces, this.City) && !lists.ContainsString(pieces, strings.TrimSuffix(this.Province, "市")) {
		pieces = append(pieces, this.City)
	}
	return strings.Join(pieces, " ")
}
