package db2struct

import (
	"bytes"
	"fmt"
	"go/format"
	"html/template"
	"io/ioutil"
	"log"
	"os"
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

// 写入不同目录的文件中(分层)
func Generate(columnTypes map[string]map[string]string, tableName string, structName string, pkgName string, jsonAnnotation bool, gormAnnotation bool, gureguTypes bool, createdKey, updatedKey string) ([]byte, error) {
	var dbTypes string
	dbTypes = generateMysqlTypes(columnTypes, 0, jsonAnnotation, gormAnnotation, gureguTypes)
	// package
	src := fmt.Sprintf("package %s", pkgName)
	// import
	src = fmt.Sprintf("%s\n%s", src, generateAllImport())
	// type struct
	src = fmt.Sprintf("%s\ntype %s %s}",
		src,
		structName,
		dbTypes)
	if gormAnnotation == true {
		// model
		if _, err := os.Stat("./model"); os.IsNotExist(err) {
			err := os.Mkdir("model",0755)
			if err != nil{
				log.Fatal(err.Error())
			}
		}
		formatted, err := format.Source([]byte(src))
		if err != nil {
			log.Fatalf("error formatting: %s, was formatting\n%s", err, src)
		}
		path := fmt.Sprintf("model/%s_model.go", tableName)
		_ = ioutil.WriteFile(path, formatted, 0644)
		// repository_interface
		if _, err := os.Stat("./repository"); os.IsNotExist(err) {
			err := os.Mkdir("repository",0755)
			if err != nil{
				log.Fatal(err.Error())
			}
		}
		src := fmt.Sprintf("package %s", "repository")
		src = fmt.Sprintf("%s\n%s", src, repoInterfaceTpl(structName, createdKey, updatedKey, tableName))
		formatted2, err := format.Source([]byte(src))
		if err != nil {
			log.Fatalf("error formatting: %s, was formatting\n%s", err, src)
		}
		path2 := fmt.Sprintf("repository/%s_repository.go", tableName)
		_ = ioutil.WriteFile(path2, formatted2, 0644)

		// repository
		if _, err := os.Stat("./repository/mysql"); os.IsNotExist(err) {
			err := os.Mkdir("repository/mysql",0755)
			if err != nil{
				log.Fatal(err.Error())
			}
		}
		src2 := fmt.Sprintf("package %s", "mysql")
		src2 = fmt.Sprintf("%s\n%s", src2, generateImport())
		src2 = fmt.Sprintf("%s\n%s", src2, repoTpl(structName, createdKey, updatedKey, tableName))
		formatted3, err := format.Source([]byte(src2))
		if err != nil {
			log.Fatalf("error formatting: %s, was formatting\n%s", err, src2)
		}
		path3 := fmt.Sprintf("repository/mysql/%s_repository.go", tableName)
		_ = ioutil.WriteFile(path3, formatted3, 0644)
	}
	return []byte("done"), nil
}

// Generate Given a Column map with datatypes and a name structName,
// attempts to generate a struct definition
// 写入一个文件
func GenerateOne(columnTypes map[string]map[string]string, tableName string, structName string, pkgName string, jsonAnnotation bool, gormAnnotation bool, gureguTypes bool, createdKey, updatedKey string) ([]byte, error) {
	var dbTypes string
	dbTypes = generateMysqlTypes(columnTypes, 0, jsonAnnotation, gormAnnotation, gureguTypes)
	// package
	src := fmt.Sprintf("package %s", pkgName)
	// import
	src = fmt.Sprintf("%s\n%s", src, generateImport())
	// type struct
	src = fmt.Sprintf("%s\ntype %s %s}",
		src,
		structName,
		dbTypes)
	if gormAnnotation == true {
		// 把所有的写入到一个文件
		src = fmt.Sprintf("%s\n%s", src, tpl(structName, createdKey, updatedKey, tableName))
		fp, _ := os.Create(tableName + ".go")
		formatted, err := format.Source([]byte(src))
		if err != nil {
			err = fmt.Errorf("error formatting: %s, was formatting\n%s", err, src)
		}
		_, _ = fp.WriteString(string(formatted))
		_ = fp.Close()
	}
	return []byte("done"), nil
}

func repoTpl(structName, createdKey, updatedKey, tableName string) string {
	t := template.New("fieldname example")
	t = t.Funcs(template.FuncMap{"lcfirst": Lcfirst})
	t = t.Funcs(template.FuncMap{"goformat": goFormat})
	t, _ = t.Parse(getRepositoryTpl())
	if createdKey == "" {
		createdKey = createdAtKey
	}
	if updatedKey == "" {
		updatedKey = updatedATKey
	}

	if createdKey == "" || updatedKey == "" {
		log.Fatal("未找到创建时间字段、更新时间字段，请指定--created_at --updated_at选项")
	}

	var buf bytes.Buffer
	var p = struct {
		StructName   string
		PrimaryKey   string
		CreatedAtKey string
		UpdatedAtKey string
		TableName    string
	}{
		structName,
		pk,
		createdKey,
		updatedKey,
		tableName,
	}
	_ = t.Execute(&buf, p)
	return buf.String()
}

func repoInterfaceTpl(structName, createdKey, updatedKey, tableName string) string {
	t := template.New("fieldname example")
	t = t.Funcs(template.FuncMap{"lcfirst": Lcfirst})
	t = t.Funcs(template.FuncMap{"goformat": goFormat})
	t, _ = t.Parse(getRepositoryInterfaceTpl())
	if createdKey == "" {
		createdKey = createdAtKey
	}
	if updatedKey == "" {
		updatedKey = updatedATKey
	}

	if createdKey == "" || updatedKey == "" {
		log.Fatal("未找到创建时间字段、更新时间字段，请指定--created_at --updated_at选项")
	}

	var buf bytes.Buffer
	var p = struct {
		StructName   string
		PrimaryKey   string
		CreatedAtKey string
		UpdatedAtKey string
		TableName    string
	}{
		structName,
		pk,
		createdKey,
		updatedKey,
		tableName,
	}
	_ = t.Execute(&buf, p)
	return buf.String()
}

func tpl(structName, createdKey, updatedKey, tableName string) string {
	t := template.New("fieldname example")
	t = t.Funcs(template.FuncMap{"lcfirst": Lcfirst})
	t = t.Funcs(template.FuncMap{"goformat": goFormat})
	t, _ = t.Parse(getTpl())
	if createdKey == "" {
		createdKey = createdAtKey
	}
	if updatedKey == "" {
		updatedKey = updatedATKey
	}

	if createdKey == "" || updatedKey == "" {
		log.Fatal("未找到创建时间字段、更新时间字段，请指定--created_at --updated_at选项")
	}

	var buf bytes.Buffer
	var p = struct {
		StructName   string
		PrimaryKey   string
		CreatedAtKey string
		UpdatedAtKey string
		TableName    string
	}{
		structName,
		pk,
		createdKey,
		updatedKey,
		tableName,
	}
	_ = t.Execute(&buf, p)
	return buf.String()
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

func FmtFieldName(s string) string {
	return fmtFieldName(s)
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
