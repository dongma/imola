package errhdl

import (
	"imola"
	"net/http"
	"testing"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	builder := NewMiddlewareBuilder()
	builder.AddCode(http.StatusNotFound, []byte(`
		<html>
			<body>
            	<h1>NOT found</h1>  
			</body>
		</html>
	`)).AddCode(http.StatusBadRequest, []byte(`
		<html>
			<body>
            	<h1>请求不对</h1>  
			</body>
		</html>
	`))
	sever := imola.NewHTTPServer(imola.ServerWithMiddleware(builder.Build()))
	sever.Start(":8081")
}
