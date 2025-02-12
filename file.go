package imola

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type FileUploader struct {
	FileField string
	// 用于计算文件的路径
	DstPathFunc func(fh *multipart.FileHeader) string
}

// Handle 文件上传功能的第一种设计实现，优势：支持额外的字段检测
func (f *FileUploader) Handle() HandleFunc {
	// 这里可做一些额外的检测，～～
	return func(ctx *Context) {
		src, srcHeader, err := ctx.Req.FormFile(f.FileField)
		if err != nil {
			ctx.RespStatusCode = http.StatusBadRequest
			ctx.RespData = []byte("上传失败，未找到数据")
			return
		}
		defer src.Close()

		dst, err := os.OpenFile(f.DstPathFunc(srcHeader), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			ctx.RespData = []byte("上传失败")
			return
		}
		defer src.Close()

		_, err = io.CopyBuffer(dst, src, nil)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			ctx.RespData = []byte("上传,拷贝文件失败")
			return
		}
		defer src.Close()
		ctx.RespData = []byte("上传成功")
	}
}

// HandleFunc Deprecated 文件上传功能的第二种实现，可直接用来注册路由，此外和Option模式配合的很好
func (f *FileUploader) HandleFunc(ctx *Context) {
	src, srcHeader, err := ctx.Req.FormFile(f.FileField)
	if err != nil {
		ctx.RespStatusCode = http.StatusBadRequest
		ctx.RespData = []byte("上传失败，未找到数据")
		return
	}
	defer src.Close()

	dst, err := os.OpenFile(f.DstPathFunc(srcHeader), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		ctx.RespStatusCode = http.StatusInternalServerError
		ctx.RespData = []byte("上传失败")
		return
	}
	defer src.Close()

	_, err = io.CopyBuffer(dst, src, nil)
	if err != nil {
		ctx.RespStatusCode = http.StatusInternalServerError
		ctx.RespData = []byte("上传,拷贝文件失败")
		return
	}
	defer src.Close()
	ctx.RespData = []byte("上传成功")
}

// FileDownloader FileDownloader直接操作http.ResponseWriter，因而middleware不能直接使用RespData
type FileDownloader struct {
	Dir string
}

// Handle 处理文件下载
func (f *FileDownloader) Handle() HandleFunc {
	return func(ctx *Context) {
		req, _ := ctx.QueryValue("file").String()
		path := filepath.Join(f.Dir, filepath.Clean(req))
		fn := filepath.Base(path)
		header := ctx.Resp.Header()
		header.Set("Content-Disposition", "attachment;filename="+fn)
		header.Set("Content-Description", "File Transfer")
		header.Set("Content-Type", "application/octet-stream")
		header.Set("Content-Transfer-Encoding", "binary")
		header.Set("Expires", "0")
		header.Set("Cache-Control", "must-revalidate")
		header.Set("Pragma", "public")
		http.ServeFile(ctx.Resp, ctx.Req, path)
	}
}

type StaticResourceHandler struct {
	dir                     string
	extensionContentTypeMap map[string]string
	// 缓存静态资源的限制
	cache       *lru.Cache
	maxFileSize int
}

// Handle 处理静态资源，包括缓存文件操作
func (h *StaticResourceHandler) Handle(ctx *Context) {
	req, _ := ctx.PathValue("file").String()
	if item, ok := h.readFileFromData(req); ok {
		log.Printf("Handle 从缓存中读数据....")
		h.writeItemAsResponse(item, ctx.Resp)
		return
	}

	path := filepath.Join(h.dir, req)
	file, err := os.Open(path)
	if err != nil {
		ctx.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	ext := getFileExt(file.Name())
	cType, ok := h.extensionContentTypeMap[ext]
	if !ok {
		ctx.Resp.WriteHeader(http.StatusBadRequest)
		return
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		ctx.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	item := &fileCacheItem{
		fileSize:    len(data),
		data:        data,
		contentType: cType,
		fileName:    req,
	}
	h.cacheFile(item)
	h.writeItemAsResponse(item, ctx.Resp)
}

// readFileFromData 从cache中根据文件名读数据
func (h *StaticResourceHandler) readFileFromData(fileName string) (*fileCacheItem, bool) {
	if h.cache != nil {
		if item, ok := h.cache.Get(fileName); ok {
			return item.(*fileCacheItem), true
		}
	}
	return nil, false
}

// writeItemAsResponse 将静态文件写到响应中
func (h *StaticResourceHandler) writeItemAsResponse(item *fileCacheItem, writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", item.contentType)
	writer.Header().Set("Content-Length", fmt.Sprintf("%d", item.fileSize))
	_, _ = writer.Write(item.data)
}

// cacheFile 缓存静态资源文件
func (h *StaticResourceHandler) cacheFile(item *fileCacheItem) {
	if h.cache != nil && item.fileSize < h.maxFileSize {
		h.cache.Add(item.fileName, item)
	}
}

type fileCacheItem struct {
	fileName    string
	fileSize    int
	contentType string
	data        []byte
}

type StaticResourceHandlerOption func(h *StaticResourceHandler)

func NewStaticResourceHandler(dir string, pathPrefix string,
	options ...StaticResourceHandlerOption) *StaticResourceHandler {
	resource := &StaticResourceHandler{
		dir: dir,
		extensionContentTypeMap: map[string]string{
			// 可根据自己的需要不断添加
			"jpeg": "image/jpeg",
			"jpe":  "image/jpeg",
			"jpg":  "image/jpeg",
			"png":  "image/png",
			"pdf":  "image/pdf",
		},
	}
	for _, opt := range options {
		opt(resource)
	}
	return resource
}

// WithFileCache 对静态文件进行缓存，maxFileSizeThreshold 缓存文件最大值（超过此大小将不会缓存）
// maxCacheFileCnt 为最多缓存文件的数量
func WithFileCache(maxFileSizeThreshold int, maxCacheFileCnt int) StaticResourceHandlerOption {
	return func(h *StaticResourceHandler) {
		cache, err := lru.New(maxCacheFileCnt)
		if err != nil {
			log.Printf("创建缓存失败，将不会缓存静态资源")
		}
		h.maxFileSize = maxFileSizeThreshold
		h.cache = cache
	}
}

// WithMoreExtension 支持更多的文件类型
func WithMoreExtension(extMap map[string]string) StaticResourceHandlerOption {
	return func(h *StaticResourceHandler) {
		for ext, contentType := range extMap {
			h.extensionContentTypeMap[ext] = contentType
		}
	}
}

// 获取文件后缀
func getFileExt(name string) string {
	index := strings.LastIndex(name, ".")
	if index == len(name)-1 {
		return ""
	}
	return name[index+1:]
}
