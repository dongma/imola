package test

import (
	"bytes"
	"html/template"
	"imola"
	"mime/multipart"
	"net/http"
	"path"
	"testing"
)

func TestFileUploader_Handle(t *testing.T) {
	server := imola.NewHTTPServer()
	server.GET("/upload_page", func(ctx *imola.Context) {
		tpl := template.New("download")
		tpl, err := tpl.Parse(`
<html>
<body>
	<form action="/download" method="post" enctype="multipart/form-data">
		 <input type="file" name="myfile" />
		 <button type="submit">上传</button>
	</form>
</body>
<html>					
`)
		if err != nil {
			t.Fatal(err)
		}
		page := &bytes.Buffer{}
		err = tpl.Execute(page, nil)
		if err != nil {
			t.Fatal(err)
		}
		ctx.RespStatusCode = http.StatusOK
		ctx.RespData = page.Bytes()
	})

	server.POST("/download", (&imola.FileUploader{
		FileField: "myFile",
		DstPathFunc: func(fh *multipart.FileHeader) string {
			return path.Join("testdata", "download", fh.Filename)
		},
	}).Handle())
	server.Start(":8081")
}

func TestFileDownloader_Handler(t *testing.T) {
	server := imola.NewHTTPServer()
	server.GET("/download", (&imola.FileDownloader{
		Dir: "./testdata/download",
	}).Handle())
	server.Start(":8081")
}

func TestStaticResource_Handler(t *testing.T) {
	server := imola.NewHTTPServer()
	handler := imola.NewStaticResourceHandler("./testdata/img", "jpeg")
	server.GET("/img/:file", handler.Handle)
	// 在浏览器中输入 http://localhost:8081/img/come_on_baby.jpg
	server.Start(":8081")
}
