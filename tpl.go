package db2struct

import "unicode"

// StructName   string
// PrimaryKey   string
// CreatedAtKey string
// UpdatedAtKey string
// TableName    string
func getTpl() string {
	return `
func (a *{{.StructName}}) TableName() string {
	return   "{{.TableName}}"
}
`
}

func Lcfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

func goFormat(str string) string {
	return fmtFieldName(stringifyFirstChar(str))
}
