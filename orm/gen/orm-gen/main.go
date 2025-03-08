package orm_gen

import (
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 用户必须输入一个src, 限制文件，然后我们会在同目录下生成代码
	src := os.Args[1]
	dstDir := filepath.Dir(src)
	fileName := filepath.Base(src)
	idx := strings.LastIndexByte(fileName, '.')
	dst := filepath.Join(dstDir, fileName[idx+1:]+".gen.go")
	f, err := os.Create(dst)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = gen(f, src)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("生成成功")
}

// Go会读取 tpl.gohtml里面的内容填充到变量tpl里面
// go:embed tpl.gohtml
var genOrm string

func gen(writer io.Writer, srcFile string) error {
	fset := token.NewFileSet()
	parsedf, err := parser.ParseFile(fset, srcFile, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	tv := &SingleFileEntryVisitor{}
	ast.Walk(tv, parsedf)
	file := tv.Get()

	tpl := template.New("gen_orm")
	tpl, err = tpl.Parse(genOrm)
	if err != nil {
		return err
	}
	return tpl.Execute(writer, OrmFile{
		File: file,
		Ops:  []string{"LT", "GT", "EQ"},
	})
}

type OrmFile struct {
	File
	Ops []string
}
