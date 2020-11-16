package setup

type SQLTable struct {
	Name       string       `json:"name"`
	Engine     string       `json:"engine"`
	Charset    string       `json:"charset"`
	Definition string       `json:"definition"`
	Fields     []*SQLField  `json:"fields"`
	Indexes    []*SQLIndex  `json:"indexes"`
	Records    []*SQLRecord `json:"records"`
}

func (this *SQLTable) FindField(fieldName string) *SQLField {
	for _, field := range this.Fields {
		if field.Name == fieldName {
			return field
		}
	}
	return nil
}

func (this *SQLTable) FindIndex(indexName string) *SQLIndex {
	for _, index := range this.Indexes {
		if index.Name == indexName {
			return index
		}
	}
	return nil
}

func (this *SQLTable) FindRecord(id int64) *SQLRecord {
	for _, record := range this.Records {
		if record.Id == id {
			return record
		}
	}
	return nil
}
