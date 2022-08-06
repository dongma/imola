package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"unsafe"
)

func main() {
	var variable int // 若未显示对变量variable赋予初值，GO编译器会为变量自动赋予类型初值
	fmt.Println("variable: ", variable)

	// 使用变量声明块(block)，可用一个var关键字将多个变量声明放在一起
	var (
		aint  int    = 128
		bint8 int8   = 6
		str   string = "hello"
		char  rune   = 'A'
		fbool bool   = true
	)
	fmt.Printf("a=%v, bint8=%v, str=%v, char=%v, f=%v \n", aint, bint8, str, char, fbool)

	// 可用一个var，声明多个变量值，当变量value未指定int类型时，go compiler会进行自动推断
	//var a, b, c int = 5, 6, 7
	//var value = 13

	var array = [6]int{1, 2, 3, 4, 5, 6}
	// length: 6, size: 68 (8*6=48), 64位平台上，int类型的大小为8，array一共有6个元素，总共48个字节
	fmt.Printf("length of array: %v, size of array: %v \n", len(array), unsafe.Sizeof(array))

	// 用go#http client启动服务，同时监听暴露的8080端口
	fmt.Println("Start go simple http server, service port: 8080")
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello world!"))
	})
	// http监听8080端口，请求时返回"hello world!"
	http.ListenAndServe(":8080", nil)
	logrus.Println("Shutdown go server service, forbid to send request!")
}
