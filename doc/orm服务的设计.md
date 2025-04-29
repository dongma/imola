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
## 数据库连接及结果集处理
go的`sql`包支持连接数据库，使用api为`sql.Open(driver, dsn)`，在使用时，不要忘记引用匿名包。对于增删改可调用`Exec`或`ExecContext`方法，其中`ExecContext`支持设置超时时间。
```go
import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // 这个不要忘记
	"imola/orm"
)
func main() {
    db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
}
```
使用反射处理结果集，`sql`查询结果用`Row`(单行数据)和`Rows`(多行数据)表示。`rows.Columns()`可用于获取数据列。
- `reflect.New(field.Typ)`能得到字段类型对应的指针，例如 `int` -> `*int`，`val.Elem()`为指针指向的类型
- vals和valElems分别保存同一个字段对应的指针和类型。因为`rows.Scan(vals...)`接受的入参是指针，只要将结果字段值写到指针中（操作内存地址），那对应的valElem字段就有值了。
- `tpValueElem := r.val`为reflect.Value，可通过`FieldByName(field.GoName).Set(valElems[i])`能将字段值通过反射写入到`struct`中。
```go
func (r reflectValue) SetColumn(rows *sql.Rows) error {
	// 拿到select出来的列
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	// 如何利用cols来解决顺序和类型问题
	vals := make([]any, 0, len(cols))
	valElems := make([]reflect.Value, 0, len(cols))
	for _, col := range cols {
		// first_name->FirstName，拿到表字段对应go struct字段
		field, ok := r.model.ColumnMap[col]
		if !ok {
			return errs.NewErrUnknownColumn(col)
		}
		// 用反射创建一个实例 (原本类型的指针类型), 例如: fd.type = int, 那么val是*int
		val := reflect.New(field.Typ)
		vals = append(vals, val.Interface())
		// 因为val是*int，那么val.Elem()的结果就是int
		valElems = append(valElems, val.Elem())
	}

	// select id, first_name, age, last_name，根据Scan的用法，其参数都是指针类型。 Scan之后，就会将sql结果写入
	err = rows.Scan(vals...)
	if err != nil {
		return err
	}

	// 将vals值塞到tp里面
	tpValueElem := r.val
	for i, col := range cols {
		field, ok := r.model.ColumnMap[col]
		if !ok {
			return errs.NewErrUnknownColumn(col)
		}
		tpValueElem.FieldByName(field.GoName).Set(valElems[i])
	}
	return err
}
```
使用`unsafe`处理结果集，`unsafe`需要掌握计算地址、计算偏移量、直接操作内存。`go`对齐规则，在`32`位机器上，按`4`个字节对齐，在`64`位机器上，按`8`个字节对齐。在大多数情况下，`unsafe`操作要比反射快，类比于`orm`之于原生`sql`查询。

`unsafe.Pointer`和`uintptr`都代表指针，那有什么区别？`unsafe.Pointer`是`Go`层面的指针，`GC`会维护`unsafe.Pointer`的值。`uintptr`直接就是一个数字，当发生`GC`时，其指向的内存地址有可能就会失效。
- 在创建`NewUnsafeValue`时，就使用`UnsafePointer()`获取了对象的起始地址。
- 在`model.go`/`Registry`注册`struct`的元数据时，在`ColumnMap`中已经解析了`struct`中每个字段类型的偏移地址。
- 在向字段设值时，通过起始地址和各个字段的不同偏移量获取字段的开始地址，然后创建字段类型指针，最后提供给`Scan`给字段赋值。
```go
func NewUnsafeValue(model *model.Model, val any) Value {
    // 实体对象的起始地址 address
    address := reflect.ValueOf(val).UnsafePointer()
    return &unsafeValue{
        model:   model,
        address: address,
    }
}

func (u unsafeValue) SetColumn(rows *sql.Rows) error {
	// 拿到select出来的列
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	var vals []any
	for _, col := range cols {
		field, ok := u.model.ColumnMap[col]
		if !ok {
			return errs.NewErrUnknownColumn(col)
		}
		// 计算字段的地址：起始地址 + 偏移量
		fdAddress := unsafe.Pointer(uintptr(u.address) + field.Offset)
		val := reflect.NewAt(field.Typ, fdAddress)
		vals = append(vals, val.Interface())
	}
	err = rows.Scan(vals...)
	return err
}
```