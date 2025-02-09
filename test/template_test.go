package test

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"html/template"
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

type FuncCall struct {
	Slice []string
}

func (f FuncCall) Hello(first string, last string) string {
	return fmt.Sprintf("%s-%s", first, last)
}
