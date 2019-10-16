package db2struct

import "unicode"

// StructName   string
// PrimaryKey   string
// CreatedAtKey string
// UpdatedAtKey string
func getTpl() string {
	return `
type {{.StructName}}Repository interface {
	Create(data *{{.StructName}}) (int, error)

	FetchOneById(id int, fields string) (*{{.StructName}}, error)
	FetchOne(where map[string]interface{}, fields string) (*{{.StructName}}, error)
	FetchByWhere(where map[string]interface{}, fields string) ([]*{{.StructName}}, error)
	FetchByIds(ids []int, fields string) ([]*{{.StructName}}, error)

	DeleteOneById(id int) error
	DeleteByWhere(where map[string]interface{}) error

	UpdateOneById(id int, set map[string]interface{}) error
	UpdateByWhere(where, set map[string]interface{}) error

	CountByWhere(where map[string]interface{}) (int, error)
	Search(where map[string]interface{}, field string, others ...map[string]interface{}) ([]*{{.StructName}}, error)
}

type {{.StructName | lcfirst }} struct {
	db *gorm.DB
}

func New{{.StructName}}(db *gorm.DB) {{.StructName}}Repository {
	return &{{.StructName | lcfirst }}{db}
}

func (a *{{.StructName|lcfirst}}) Create(data *{{.StructName}}) (int, error) {
	if a.db.NewRecord(data) {
		data.{{.CreatedAtKey|goformat}} = time.Now()
		data.{{.UpdatedAtKey|goformat}} = time.Now()
		if err := a.db.Create(data).Error; err != nil {
			return 0, err
		}
		return data.{{.PrimaryKey|goformat}}, nil
	}
	return 0, errors.New("this is not a new record")
}

func (a *{{.StructName|lcfirst}}) FetchOneById(id int, fields string) (*{{.StructName}}, error) {
	var ret {{.StructName}}

	err := a.db.Select(fields).First(&ret, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *{{.StructName|lcfirst}}) FetchOne(where map[string]interface{}, fields string) (*{{.StructName}}, error) {
	var ret {{.StructName}}

	q := a.db.Select(fields)
	for k, v := range where {
		if v != nil {
			q = q.Where(k, v)
		} else {
			q = q.Where(k)
		}
	}

	err := q.First(&ret).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (a *{{.StructName|lcfirst}}) FetchByWhere(where map[string]interface{}, fields string) ([]*{{.StructName}}, error) {
	var ret []*{{.StructName}}

	q := a.db.Select(fields)
	for k, v := range where {
		if v != nil {
			q = q.Where(k, v)
		} else {
			q = q.Where(k)
		}
	}

	err := q.Find(&ret).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *{{.StructName|lcfirst}}) FetchByIds(ids []int, fields string) ([]*{{.StructName}}, error) {
	var ret []*{{.StructName}}

	err := a.db.Select(fields).Find(&ret, ids).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *{{.StructName|lcfirst}}) DeleteOneById(id int) error {
	d := {{.StructName}}{ {{.PrimaryKey|goformat}}: id}
	if err := a.db.Delete(&d).Limit(1).Error; err != nil {
		return err
	}
	return nil
}

func (a *{{.StructName|lcfirst}}) DeleteByWhere(where map[string]interface{}) error {
	q := a.db
	for k, v := range where {
		if v != nil {
			q = q.Where(k, v)
		} else {
			q = q.Where(k)
		}
	}
	if err := q.Delete({{.StructName}}{}).Error; err != nil {
		return err
	}
	return nil
}

func (a *{{.StructName|lcfirst}}) UpdateOneById(id int, set map[string]interface{}) error {
	set["{{.UpdatedAtKey}}"] = time.Now()
	if err := a.db.Model({{.StructName}}{ {{.PrimaryKey|goformat}}: id}).Update(set).Limit(1).Error; err != nil {
		return err
	}
	return nil
}

func (a *{{.StructName|lcfirst}}) UpdateByWhere(where, set map[string]interface{}) error {
	set["{{.UpdatedAtKey}}"] = time.Now()

	q := a.db.Model({{.StructName}}{})
	for k, v := range where {
		q = q.Where(k, v)
	}

	if err := q.Update(set).Error; err != nil {
		return err
	}
	return nil
}

func (a *{{.StructName|lcfirst}}) CountByWhere(where map[string]interface{}) (int, error) {
	c := 0

	q := a.db.Model({{.StructName}}{})
	for k, v := range where {
		if v != nil {
			q = q.Where(k, v)
		} else {
			q = q.Where(k)
		}
	}
	if err := q.Count(&c).Error; err != nil {
		return 0, err
	}

	return c, nil
}

func (a *{{.StructName|lcfirst}}) Search(where map[string]interface{}, field string, others ...map[string]interface{}) ([]*{{.StructName}}, error) {
	var ret []*{{.StructName}}

	q := a.db.Select(field)
	for k, v := range where {
		if v != nil {
			q = q.Where(k, v)
		} else {
			q = q.Where(k)
		}
	}

	if others != nil {
		if g, ok := others[0]["joins"]; ok {
			for _, j := range g.([]string) {
				q = q.Joins(j)
			}
		}

		if g, ok := others[0]["group"]; ok {
			q = q.Group(g.(string))
		}

		if h, ok := others[0]["having"]; ok {
			for k, v := range h.(map[string]interface{}) {
				q = q.Having(k, v)
			}
		}

		if o, ok := others[0]["order"]; ok {
			q = q.Order(o)
		}

		if o, ok := others[0]["offset"]; ok {
			q = q.Offset(o)
		}
		if l, ok := others[0]["limit"]; ok {
			q = q.Limit(l)
		}
	}

	if err := q.Find(&ret).Error; err != nil {
		return ret, nil
	}

	return ret, nil
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
