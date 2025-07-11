package test

import (
	"bytes"
	"fmt"
	"github.com/dongma/imola/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"html/template"
	"log"
	"testing"
)

func TestHelloWorld(t *testing.T) {
	type User struct {
		Name string
	}
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`hello, {{ .Name }}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, User{Name: "Tom"})
	require.NoError(t, err)
	assert.Equal(t, `hello, Tom`, buffer.String())
}

func TestMapData(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`hello, {{ .Name }}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, map[string]string{"Name": "Tom"})
	require.NoError(t, err)
	assert.Equal(t, `hello, Tom`, buffer.String())
}

func TestSlice(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`hello, {{ index . 0}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, []string{"Tom"})
	require.NoError(t, err)
	assert.Equal(t, `hello, Tom`, buffer.String())
}

func TestBasic(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`hello, {{.}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, 123)
	require.NoError(t, err)
	assert.Equal(t, `hello, 123`, buffer.String())
}

func TestFuncCall(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`
切片长度：{{len .Slice}}
Hello, {{.Hello "Tom" "Jerry"}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, &FuncCall{
		Slice: []string{"a", "b"},
	})
	require.NoError(t, err)
	assert.Equal(t, `
切片长度：2
Hello, Tom-Jerry`, buffer.String())
}

func TestLoginPage(t *testing.T) {
	tpl, err := template.ParseGlob("testdata/tpls/*.gohtml")
	require.NoError(t, err)
	engine := &web.GoTemplateEngine{
		T: tpl,
	}

	h := web.NewHTTPServer(web.ServerWithTemplateEngine(engine))
	h.GET("/login", func(ctx *web.Context) {
		err := ctx.Render("login.gohtml", nil)
		if err != nil {
			log.Println(err)
		}
	})
	h.Start(":8081")
}

func TestIfElse(t *testing.T) {
	tpl := template.New("hello-world")
	tpl, err := tpl.Parse(`
		{{- if and (gt .Age 0) (le .Age 6)}}
		我是儿童: (0, 6]
		{{ else if and (gt .Age 6) (le .Age 18) }}
		我是少年: (6, 18]
		{{ else }}我是成人: >18{{end -}}
	`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, User{Age: 19})
	require.NoError(t, err)
	assert.Equal(t, `我是成人: >18`, buffer.String())
}

type User struct {
	Age uint32
}

type FuncCall struct {
	Slice []string
}

func (f FuncCall) Hello(first string, last string) string {
	return fmt.Sprintf("%s-%s", first, last)
}
