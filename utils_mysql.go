package db2struct

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var sortFields []string

// GetColumnsFromMysqlTable Select column details from information schema and return map of map
func GetColumnsFromMysqlTable(mariadbUser string, mariadbPassword string, mariadbHost string, mariadbPort int, mariadbDatabase string, mariadbTable string) (*map[string]map[string]string, error) {

	var err error
	var db *sql.DB
	if mariadbPassword != "" {
		db, err = sql.Open("mysql", mariadbUser+":"+mariadbPassword+"@tcp("+mariadbHost+":"+strconv.Itoa(mariadbPort)+")/"+mariadbDatabase+"?&parseTime=True")
	} else {
		db, err = sql.Open("mysql", mariadbUser+"@tcp("+mariadbHost+":"+strconv.Itoa(mariadbPort)+")/"+mariadbDatabase+"?&parseTime=True")
	}
	defer db.Close()

	// Check for error in db, note this does not check connectivity but does check uri
	if err != nil {
		fmt.Println("Error opening mysql db: " + err.Error())
		return nil, err
	}

	// Store colum as map of maps
	columnDataTypes := make(map[string]map[string]string)
	// Select columnd data from INFORMATION_SCHEMA
	columnDataTypeQuery := "SELECT COLUMN_NAME, COLUMN_KEY, DATA_TYPE, IS_NULLABLE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = ? AND table_name = ? ORDER BY ORDINAL_POSITION ASC"

	if Debug {
		fmt.Println("running: " + columnDataTypeQuery)
	}

	rows, err := db.Query(columnDataTypeQuery, mariadbDatabase, mariadbTable)

	if err != nil {
		fmt.Println("Error selecting from db: " + err.Error())
		return nil, err
	}
	if rows != nil {
		defer rows.Close()
	} else {
		return nil, errors.New("No results returned for table")
	}

	for rows.Next() {
		var column string
		var columnKey string
		var dataType string
		var nullable string
		rows.Scan(&column, &columnKey, &dataType, &nullable)

		columnDataTypes[column] = map[string]string{"value": dataType, "nullable": nullable, "primary": columnKey}
		sortFields = append(sortFields, column)
	}

	return &columnDataTypes, err
}

var pk, createdAtKey, updatedATKey string
var haveNull bool

func generateAllImport() string {
	i := `
import (
	"time"
	
	"github.com/jinzhu/gorm"
`
	if haveNull == true {
		i = fmt.Sprintf("%s\"gopkg.in/guregu/null.v3\"", i)
	}
	i += `
)
`
	return i
}

func generateImport() string {
	i := `
import (
	"errors"
	"time"
	
	"github.com/jinzhu/gorm"
`
	if haveNull == true {
		i = fmt.Sprintf("%s\"gopkg.in/guregu/null.v3\"", i)
	}
	i += `
)
`
	return i
}

// Generate go struct entries for a map[string]interface{} structure
func generateMysqlTypes(obj map[string]map[string]string, depth int, jsonAnnotation bool, gormAnnotation bool, gureguTypes bool) string {
	structure := "struct {"

	l := len(sortFields)
	for i := 0; i < l; i++ {
		key := sortFields[i]
		mysqlType := obj[key]
		nullable := false
		if mysqlType["nullable"] == "YES" {
			nullable = true
		}

		primary := ""
		if mysqlType["primary"] == "PRI" {
			pk = key
			primary = ";primary_key"
		}

		if mysqlType["value"] == "timestamp" || mysqlType["value"] == "datetime" {
			if strings.Contains(key, "create") {
				createdAtKey = key
			}
			if strings.Contains(key, "update") {
				updatedATKey = key
			}
		}

		// Get the corresponding go value type for this mysql type
		var valueType string
		// If the guregu (https://github.com/guregu/null) CLI option is passed use its types, otherwise use go's sql.NullX

		valueType = mysqlTypeToGoType(mysqlType["value"], nullable, gureguTypes)

		fieldName := fmtFieldName(stringifyFirstChar(key))
		var annotations []string
		if gormAnnotation == true {
			annotations = append(annotations, fmt.Sprintf("gorm:\"column:%s%s\"", key, primary))
		}
		if jsonAnnotation == true {
			//annotations = append(annotations, fmt.Sprintf("json:\"%s%s\"", key, primary))
			annotations = append(annotations, fmt.Sprintf("json:\"%s\"", key))
		}
		if len(annotations) > 0 {
			structure += fmt.Sprintf("\n%s %s `%s`",
				fieldName,
				valueType,
				strings.Join(annotations, " "))

		} else {
			structure += fmt.Sprintf("\n%s %s",
				fieldName,
				valueType)
		}
	}
	return structure
}

// mysqlTypeToGoType converts the mysql types to go compatible sql.Nullable (https://golang.org/pkg/database/sql/) types
func mysqlTypeToGoType(mysqlType string, nullable bool, gureguTypes bool) string {
	switch mysqlType {
	case "tinyint", "int", "smallint", "mediumint":
		if nullable {
			if gureguTypes {
				haveNull = true
				return gureguNullInt
			}
			return sqlNullInt
		}
		return golangInt
	case "bigint":
		if nullable {
			if gureguTypes {
				haveNull = true
				return gureguNullInt
			}
			return sqlNullInt
		}
		return golangInt64
	case "char", "enum", "varchar", "longtext", "mediumtext", "text", "tinytext":
		if nullable {
			if gureguTypes {
				haveNull = true
				return gureguNullString
			}
			return sqlNullString
		}
		return "string"
	case "date", "datetime", "time", "timestamp":
		if nullable && gureguTypes {
			haveNull = true
			return gureguNullTime
		}
		return golangTime
	case "decimal", "double":
		if nullable {
			if gureguTypes {
				haveNull = true
				return gureguNullFloat
			}
			return sqlNullFloat
		}
		return golangFloat64
	case "float":
		if nullable {
			if gureguTypes {
				haveNull = true
				return gureguNullFloat
			}
			return sqlNullFloat
		}
		return golangFloat32
	case "binary", "blob", "longblob", "mediumblob", "varbinary":
		return golangByteArray
	}
	return ""
}
