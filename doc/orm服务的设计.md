## ORM框架概览
`ORM`: 对象-关系映射，`ORM`框架则是帮助用户完成对象到`SQL`，以及结果集到对象的映射的工具，`ORM`功能点：
- SQL：必须要支持增删，改查，DDL一般是作为一个扩展功能，或者作为一个工具来提供；
- 映射：将结果集封装成对象，性能瓶颈，事务：主要在与维护好事务状态。
- 元数据：SQL和映射两部分的基石，AOP：处理横向关注点
- 关联关系：部分ORM框架会提供，性价比低。方言，兼容不同的数据库，至少要兼容MySQL、SQLite、PostgreSQL。

核心接口设计`Selector`先构建查询的`SQL`，mysql语法规范包括：`from`结构，`where`和`having`以及`order by`和`limit`。
```go
type Selector[T any] struct {
	builder
	table   TableReference
	where   []sql.Predicate
	columns []Selectable
}

func (s *Selector[T]) Build() (*Query, error) {
	// 用于构建select、from和where这三部分内容，参考mysql语法规范
}
```
`where`的设计需考虑谓词，左表达式 Op 右表达式，如果是`Not`，则左边缺省，只剩下`Op` `Right`。
```go
type Predicate struct {
	Left  Expression
	Op    Op
	Right Expression
}
func (Predicate) Expr() {}

// And 用法: C("id").Eq(12).And(C("Name").Eq("Tom"))
func (left Predicate) And(right Predicate) Predicate {
    return Predicate{
        Left:  left,
        Op:    OpAnd,
        Right: right,
    }
}

// Or 用法: C("id").Eq(12).Or(C("Name").Eq("Tom"))
func (left Predicate) Or(right Predicate) Predicate {
    return Predicate{
        Left:  left,
        Op:    OpOr,
        Right: right,
    }
}
```
## 元数据
ORM框架需要解析模型以获得模型的元数据，这些元数据将被用来构建SQL、执行校验、以及用于处理结果集。最简版的元数据如下`Model`和`Field`两个属性，要实现将`Model`转换为`SQL`需要用到反射机制。
```go
type Model struct {
	TableName string
	Fields    []*Field
}
type Field struct {
    // 列名-mysql的，例如：first_name, last_name的格式
    Column string
}
```
使用反射解析模型及字段，`reflect.Type`用于获取实体类型，这样可以得到一个`struct`的元数据。优化考虑，减少多次解析同一个`struct`的元数据，可以考虑元数据注册中心`Registry`，其内包含一个`sync.Map`，其key为`reflect.Type`。

在`struct`上可以通过解析标签`tag`获取字段的别名，例如 Name string `gorm:"c_name"`，可获取到表名为`c_name`。另外，在`go`中经常使用的`Option`模式，也可实现同样功能，详细内容请看`model/model.go`的`WithTableName`和`WithColumnName`方法。
```go
// Register 限制只能使用一级指针
func (r *Registry) Register(entity any, opts ...Option) (*Model, error) {
	typ := reflect.TypeOf(entity) // 获取当前entity的类型
	// typ.Kind()可用于获取类型，是否为指针？ typ.Elem().Kind()为指针指向的对象的类型
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	elemTyp := typ.Elem()
	numField := elemTyp.NumField()  // 指针指向的struct对象，含有字段的个数
	fields := make([]*Field, 0, numField)

	for i := 0; i < numField; i++ {
		fdType := elemTyp.Field(i)  // 取第i个字段的类型
		fdMeta := &Field{   // 将字段名转为：中间字母下划线格式
			Column: underscoreName(fdType.Name),
			Typ:    fdType.Type,
		}
		fields = append(fields, fdMeta)
	}

	var tableName := underscoreName(elemTyp.Name()) // 表名也转换为下划线格式，例如：user_detail
	res := &Model{
		TableName: tableName,
		Fields:    fields,
	}
	return res, nil
}

// Registry 代表元数据的注册中心
type Registry struct {
	Models sync.Map
}
```
