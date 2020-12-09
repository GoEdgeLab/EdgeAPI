package setup

import "regexp"

type SQLField struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
}

func (this *SQLField) EqualDefinition(definition2 string) bool {
	if this.Definition == definition2 {
		return true
	}

	// 针对MySQL v8.0.17以后
	def1 := regexp.MustCompile(`(?)(tinyint|smallint|mediumint|int|bigint)\(\d+\)`).
		ReplaceAllString(this.Definition, "${1}")
	return def1 == definition2
}
