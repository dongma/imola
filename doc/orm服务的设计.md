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
`where`的设计需考虑谓词，左表达式 Op 右表达式，如果是`Not`，则左边缺省，只剩下Op Right。
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
