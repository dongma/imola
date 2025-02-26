package model

import (
	"imola/orm/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

const (
	tagColumn = "column"
)

type ModelOpt func(model *Model) error

// ModelWithTableName 自定义表名，对表明没有任何校验
func ModelWithTableName(tableName string) ModelOpt {
	return func(model *Model) error {
		model.TableName = tableName
		return nil
	}
}

// ModelWithColumnName 自定义列名
func ModelWithColumnName(field string, columnName string) ModelOpt {
	return func(model *Model) error {
		fd, ok := model.FieldMap[field]
		if !ok {
			return errs.NewErrUnknownField(field)
		}
		// 并未检测columnName是否为空字符串
		fd.Column = columnName
		return nil
	}
}

// IRegistry 使用编程的方式 注册元数据
type IRegistry interface {
	// Get 查找一个模型
	Get(val any) (*Model, error)
	// Register 注册一个模型
	Register(val any, opts ...ModelOpt) (*Model, error)
}

type Model struct {
	TableName string
	// 字段名到字段定义的映射，Id->id, FirstName->first_name
	FieldMap map[string]*Field
	// 列名到字段定义的映射, id->Id, first_name->FirstName
	ColumnMap map[string]*Field
}

type Field struct {
	GoName string
	// 列名
	Column string
	// 字段的类型
	Typ reflect.Type
	// 字段相对于结构体的偏移量
	Offset uintptr
}

// Registry 代表元数据的注册中心
type Registry struct {
	//lock   sync.RWMutex
	Models sync.Map
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Get(val any) (*Model, error) {
	typ := reflect.TypeOf(val)
	m, ok := r.Models.Load(typ)
	if ok {
		return m.(*Model), nil
	}
	m, err := r.Register(val)
	if err != nil {
		return nil, err
	}
	return m.(*Model), nil
}

// Get 使用了go double-check的写法, 与使用sync.Map同类型
/*func (r *Registry) Get(val any) (*Model, error) {
	typ := reflect.TypeOf(val)
	r.lock.RLock()
	m, ok := r.Models[typ]
	r.lock.RUnlock()

	if ok {
		return m, nil
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	m, err := r.Register(val)
	if err != nil {
		return nil, err
	}
	r.Models[typ] = m
	return m, nil
}*/

// Register 限制只能使用一级指针
func (r *Registry) Register(entity any, opts ...ModelOpt) (*Model, error) {
	typ := reflect.TypeOf(entity)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	elemTyp := typ.Elem()
	numField := elemTyp.NumField()
	fieldMap := make(map[string]*Field, numField)
	columnMap := make(map[string]*Field, numField)

	for i := 0; i < numField; i++ {
		fdType := elemTyp.Field(i)
		pair, err := r.ParseTag(fdType.Tag)
		if err != nil {
			return nil, err
		}
		colName := pair[tagColumn]
		if colName == "" {
			// 用户没有设置tag别名
			colName = underscoreName(fdType.Name)
		}
		fdMeta := &Field{
			Column: colName,
			// 获取字段类型
			Typ:    fdType.Type,
			GoName: fdType.Name,
			Offset: fdType.Offset,
		}
		fieldMap[fdType.Name] = fdMeta
		columnMap[colName] = fdMeta
	}

	var tableName string
	if tbl, ok := entity.(TableName); ok {
		tableName = tbl.TableName()
	}
	if tableName == "" {
		tableName = underscoreName(elemTyp.Name())
	}
	res := &Model{
		TableName: tableName,
		FieldMap:  fieldMap,
		ColumnMap: columnMap,
	}
	for _, opt := range opts {
		err := opt(res)
		if err != nil {
			return nil, err
		}
	}
	r.Models.Store(typ, res)
	return res, nil
}

// ParseTag 标签解析策略
func (r *Registry) ParseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag := tag.Get("orm")
	// 返回一个空的map，这样调用者就不需要判断nil值了
	if ormTag == "" {
		return map[string]string{}, nil
	}
	res := make(map[string]string, 1)
	// 字符串处理
	pairs := strings.Split(ormTag, ",")
	for _, pair := range pairs {
		segs := strings.Split(pair, "=")
		if len(segs) != 2 {
			return nil, errs.NewErrInvalidTagContent(pair)
		}
		key := segs[0]
		val := segs[1]
		res[key] = val
	}
	return res, nil
}

// underscoreName 驼峰转换字符串
func underscoreName(tableName string) string {
	var buf []byte
	for i, v := range tableName {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}
	}
	return string(buf)
}

// TableName 用户实现这个接口来返回自定义的表名
type TableName interface {
	TableName() string
}
