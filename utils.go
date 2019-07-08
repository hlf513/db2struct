package db2struct

import (
	"fmt"
	"go/format"
	"strconv"
	"strings"
	"unicode"
)

// Constants for return types of golang
const (
	golangByteArray  = "[]byte"
	gureguNullInt    = "null.Int"
	sqlNullInt       = "sql.NullInt64"
	golangInt        = "int"
	golangInt64      = "int64"
	gureguNullFloat  = "null.Float"
	sqlNullFloat     = "sql.NullFloat64"
	golangFloat      = "float"
	golangFloat32    = "float32"
	golangFloat64    = "float64"
	gureguNullString = "null.String"
	sqlNullString    = "sql.NullString"
	gureguNullTime   = "null.Time"
	golangTime       = "time.Time"
)

// commonInitialisms is a set of common initialisms.
// Only add entries that are highly unlikely to be non-initialisms.
// For instance, "ID" is fine (Freudian code is rare), but "AND" is not.
var commonInitialisms = map[string]bool{
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SSH":   true,
	"TLS":   true,
	"TTL":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
}

var intToWordMap = []string{
	"zero",
	"one",
	"two",
	"three",
	"four",
	"five",
	"six",
	"seven",
	"eight",
	"nine",
}

//Debug level logging
var Debug = false

// Generate Given a Column map with datatypes and a name structName,
// attempts to generate a struct definition
func Generate(columnTypes map[string]map[string]string, tableName string, structName string, pkgName string, jsonAnnotation bool, gormAnnotation bool, gureguTypes bool, primaryKey, createdKey, updatedKey, dbModel string) ([]byte, error) {
	var dbTypes string
	dbTypes = generateMysqlTypes(columnTypes, 0, jsonAnnotation, gormAnnotation, gureguTypes)
	src := fmt.Sprintf("package %s\ntype %s %s}",
		pkgName,
		structName,
		dbTypes)
	if gormAnnotation == true {
		tableNameFunc := "// TableName sets the insert table name for this struct type\n" +
			"func (" + strings.ToLower(string(structName[0])) + " *" + structName + ") TableName() string {\n" +
			"	return \"" + tableName + "\"" +
			"}"
		src = fmt.Sprintf("%s\n%s", src, tableNameFunc)
		src = fmt.Sprintf("%s\n%s", src, tpl(structName, primaryKey, createdKey, updatedKey, dbModel))
	}
	formatted, err := format.Source([]byte(src))
	if err != nil {
		err = fmt.Errorf("error formatting: %s, was formatting\n%s", err, src)
	}
	return formatted, err
}

func tpl(structName, primaryKey, createdKey, updatedKey, dbModel string) string {
	tpl := "func(" + strings.ToLower(string(structName[0])) + " *" + structName + ") Create(data *" + structName + ") (int, error) {\n" +
		"	if "+ dbModel +".NewRecord(data) {\n" +
		"		data." + createdKey + "= time.Now() \n" +
		"		data." + updatedKey + "= time.Now()	\n" +
		"		if err := "+ dbModel +".Create(data).Error; err != nil {\n" +
		"			return 0, err\n" +
		"		}\n" +
		"		return data." + primaryKey + ", nil\n" +
		"	}\n" +
		"	return 0, errors.New(\"this is not a new record\")\n" +
		"}\n" +
		"\n" +
		"func(" + strings.ToLower(string(structName[0])) + " *" + structName + ") FetchOneById(id int, fields string) (" + structName + ", error) {\n" +
		"	var ret " + structName + "\n" +
		"\n" +
		"	if err := "+ dbModel +".Select(fields).First(&ret, id).Error; err != nil {\n" +
		"		return ret, err\n" +
		"	}\n" +
		"\n" +
		"	return ret, nil\n" +
		"}\n" +
		"\n" +
		"func(" + strings.ToLower(string(structName[0])) + " *" + structName + ") FetchOne(where map[string]interface{}, fields string) (" + structName + ", error) {\n" +
		"	var ret " + structName + "\n" +
		"\n" +
		"	q := "+ dbModel +".Select(fields)\n" +
		"	for k, v := range where {\n" +
		"		q = q.Where(k, v)\n" +
		"	}\n" +
		"\n" +
		"	if err := q.First(&ret).Error; err != nil {\n" +
		"		return ret, err\n" +
		"	}\n" +
		"\n" +
		"	return ret, nil\n" +
		"}\n" +
		"\n" +
		"func(" + strings.ToLower(string(structName[0])) + " *" + structName + ") FetchByWhere(where map[string]interface{}, fields string) ([]" + structName + ", error) {\n" +
		"	var ret []" + structName + "\n" +
		"\n" +
		"	q := "+ dbModel +".Select(fields)\n" +
		"	for k, v := range where {\n" +
		"		q = q.Where(k, v)\n" +
		"	}\n" +
		"\n" +
		"	if err := q.Find(&ret).Error; err != nil {\n" +
		"		return ret, err\n" +
		"	}\n" +
		"\n" +
		"	return ret, nil\n" +
		"}\n" +
		"\n" +
		"func(" + strings.ToLower(string(structName[0])) + " *" + structName + ") FetchByIds(ids []int, fields string) ([]" + structName + ", error) {\n" +
		"	var ret []" + structName + "\n" +
		"\n" +
		"	if err := "+ dbModel +".Select(fields).Find(&ret, ids).Error; err != nil {\n" +
		"		return ret, err\n" +
		"	}\n" +
		"	return ret, nil\n" +
		"}\n" +
		"\n" +
		"func(" + strings.ToLower(string(structName[0])) + " *" + structName + ") DeleteOneById(id int) error {\n" +
		"	d := " + structName + "{" + primaryKey + ": id}\n" +
		"	if err := "+ dbModel +".Delete(&d).Limit(1).Error; err != nil {\n" +
		"		return err\n" +
		"	}\n" +
		"	return nil\n" +
		"}\n" +
		"\n" +
		"func(" + strings.ToLower(string(structName[0])) + " *" + structName + ") DeleteByWhere(where map[string]interface{}) error {\n" +
		" 	q := " + dbModel + "\n" +
		" 	for k, v := range where {\n" +
		" 		q = q.Where(k, v)\n" +
		" 	}\n" +
		"	if err := q.Delete(" + strings.ToLower(string(structName[0])) + ").Error; err != nil {\n" +
		"		return err\n" +
		"	}\n" +
		"	return nil\n" +
		"}\n" +
		"\n" +
		"func(" + strings.ToLower(string(structName[0])) + " *" + structName + ") UpdateOneById(id int, set map[string]interface{}) error {\n" +
		"	set[\"" + updatedKey + "\"] = time.Now()\n" +
		"	if err := "+ dbModel +".Model(" + structName + "{" + primaryKey + ": id}).Update(set).Limit(1).Error; err != nil {\n" +
		"		return err\n" +
		"	}\n" +
		"	return nil\n" +
		"}\n" +
		"\n" +
		"func(" + strings.ToLower(string(structName[0])) + " *" + structName + ") UpdateByWhere(where, set map[string]interface{}) error {\n" +
		"	set[\"" + updatedKey + "\"] = time.Now()\n" +
		"	q := "+ dbModel +".Model(" + strings.ToLower(string(structName[0])) + ")\n" +
		"	for k, v := range where {\n" +
		"		q = q.Where(k, v)\n" +
		"	}\n" +
		"\n" +
		"	if err := q.Update(set).Error; err != nil {\n" +
		"		return err\n" +
		"	}\n" +
		"	return nil\n" +
		"}\n" +
		"\n" +
		"func(" + strings.ToLower(string(structName[0])) + " *" + structName + ") CountByWhere(where map[string]interface{}) (int, error) {\n" +
		"	c := 0\n" +
		"\n" +
		"	q := "+ dbModel +".Model(" + strings.ToLower(string(structName[0])) + ")\n" +
		"	for k, v := range where {\n" +
		"		q = q.Where(k, v)\n" +
		"	}\n" +
		"	if err := q.Count(&c).Error; err != nil {\n" +
		"		return 0, err\n" +
		"	}\n" +
		"\n" +
		"	return c, nil\n" +
		"}\n" +
		"\n" +
		"func(" + strings.ToLower(string(structName[0])) + " *" + structName + ") Search(where map[string]interface{}, field string, others ...map[string]interface{}) ([]" + structName + ", error) {\n" +
		"	var ret []" + structName + "\n" +
		"\n" +
		"	q := "+ dbModel +".Select(field)\n" +
		"	for k, v := range where {\n" +
		"		q = q.Where(k, v)\n" +
		"	}\n" +
		"\n" +
		"	if others != nil {\n" +
		"		if g, ok := others[0][\"group\"]; ok {\n" +
		"			q = q.Group(g.(string))\n" +
		"		}\n" +
		"\n" +
		"		if h, ok := others[0][\"having\"]; ok {\n" +
		"			for k, v := range h.(map[string]interface{}) {\n" +
		"				q = q.Having(k, v)\n" +
		"			}\n" +
		"		}\n" +
		"\n" +
		"		if o, ok := others[0][\"order\"]; ok {\n" +
		"			q = q.Order(o)\n" +
		"		}\n" +
		"\n" +
		"		if o, ok := others[0][\"offset\"]; ok {\n" +
		"			q = q.Offset(o)\n" +
		"		}\n" +
		"		if l, ok := others[0][\"limit\"]; ok {\n" +
		"			q = q.Limit(l)\n" +
		"		}\n" +
		"	}\n" +
		"\n" +
		"	if err := q.Find(&ret).Error; err != nil {\n" +
		"		return ret, nil\n" +
		"	}\n" +
		"\n" +
		"	return ret, nil\n" +
		"}\n"

	return tpl
}

// fmtFieldName formats a string as a struct key
//
// Example:
// 	fmtFieldName("foo_id")
// Output: FooID
func fmtFieldName(s string) string {
	name := lintFieldName(s)
	runes := []rune(name)
	for i, c := range runes {
		ok := unicode.IsLetter(c) || unicode.IsDigit(c)
		if i == 0 {
			ok = unicode.IsLetter(c)
		}
		if !ok {
			runes[i] = '_'
		}
	}
	return string(runes)
}

func lintFieldName(name string) string {
	// Fast path for simple cases: "_" and all lowercase.
	if name == "_" {
		return name
	}

	for len(name) > 0 && name[0] == '_' {
		name = name[1:]
	}

	allLower := true
	for _, r := range name {
		if !unicode.IsLower(r) {
			allLower = false
			break
		}
	}
	if allLower {
		runes := []rune(name)
		if u := strings.ToUpper(name); commonInitialisms[u] {
			copy(runes[0:], []rune(u))
		} else {
			runes[0] = unicode.ToUpper(runes[0])
		}
		return string(runes)
	}

	// Split camelCase at any lower->upper transition, and split on underscores.
	// Check each word for common initialisms.
	runes := []rune(name)
	w, i := 0, 0 // index of start of word, scan
	for i+1 <= len(runes) {
		eow := false // whether we hit the end of a word

		if i+1 == len(runes) {
			eow = true
		} else if runes[i+1] == '_' {
			// underscore; shift the remainder forward over any run of underscores
			eow = true
			n := 1
			for i+n+1 < len(runes) && runes[i+n+1] == '_' {
				n++
			}

			// Leave at most one underscore if the underscore is between two digits
			if i+n+1 < len(runes) && unicode.IsDigit(runes[i]) && unicode.IsDigit(runes[i+n+1]) {
				n--
			}

			copy(runes[i+1:], runes[i+n+1:])
			runes = runes[:len(runes)-n]
		} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
			// lower->non-lower
			eow = true
		}
		i++
		if !eow {
			continue
		}

		// [w,i) is a word.
		word := string(runes[w:i])
		if u := strings.ToUpper(word); commonInitialisms[u] {
			// All the common initialisms are ASCII,
			// so we can replace the bytes exactly.
			copy(runes[w:], []rune(u))

		} else if strings.ToLower(word) == word {
			// already all lowercase, and not the first word, so uppercase the first character.
			runes[w] = unicode.ToUpper(runes[w])
		}
		w = i
	}
	return string(runes)
}

// convert first character ints to strings
func stringifyFirstChar(str string) string {
	first := str[:1]

	i, err := strconv.ParseInt(first, 10, 8)

	if err != nil {
		return str
	}

	return intToWordMap[i] + "_" + str[1:]
}
