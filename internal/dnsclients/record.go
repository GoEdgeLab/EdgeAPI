package dnsclients

type RecordType = string

const (
	RecordTypeA     RecordType = "A"
	RecordTypeCName RecordType = "CNAME"
	RecordTypeTXT   RecordType = "TXT"
)

type Record struct {
	Id    string     `json:"id"`
	Name  string     `json:"name"`
	Type  RecordType `json:"type"`
	Value string     `json:"value"`
	Route string     `json:"route"`
}
