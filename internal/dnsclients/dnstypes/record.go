package dnstypes

type RecordType = string

const (
	RecordTypeA     RecordType = "A"
	RecordTypeAAAA  RecordType = "AAAA"
	RecordTypeCNAME RecordType = "CNAME"
	RecordTypeTXT   RecordType = "TXT"
)

type Record struct {
	Id    string     `json:"id"`
	Name  string     `json:"name"`
	Type  RecordType `json:"type"`
	Value string     `json:"value"`
	Route string     `json:"route"`
	TTL   int32      `json:"ttl"`
}

func (this *Record) Clone() *Record {
	return &Record{
		Id:    this.Id,
		Name:  this.Name,
		Type:  this.Type,
		Value: this.Value,
		Route: this.Route,
		TTL:   this.TTL,
	}
}

func (this *Record) Copy(anotherRecord *Record) {
	if anotherRecord == nil {
		return
	}
	this.Id = anotherRecord.Id
	this.Name = anotherRecord.Name
	this.Type = anotherRecord.Type
	this.Value = anotherRecord.Value
	this.Route = anotherRecord.Route
	this.TTL = anotherRecord.TTL
}
