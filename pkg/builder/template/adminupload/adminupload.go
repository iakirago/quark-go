package adminupload

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"

	"github.com/quarkcms/quark-go/pkg/builder"
	"github.com/quarkcms/quark-go/pkg/builder/template"
	"github.com/quarkcms/quark-go/pkg/dal/db"
	"github.com/quarkcms/quark-go/pkg/msg"
	"github.com/quarkcms/quark-go/pkg/storage"
)

// 文件上传
type Template struct {
	template.AdminTemplate
	LimitSize        int64              // 限制文件大小
	LimitType        interface{}        // 限制文件类型
	LimitImageWidth  int64              // 限制图片宽度
	LimitImageHeight int64              // 限制图片高度
	Driver           string             // 存储驱动
	SavePath         string             // 保存路径
	OSSConfig        *storage.OSSConfig // OSS配置
}

// 初始化
func (p *Template) Init() interface{} {
	p.TemplateInit()

	return p
}

// 初始化模板
func (p *Template) TemplateInit() interface{} {

	// 初始化数据对象
	p.DB = db.Client

	// 默认本地上传
	p.Driver = storage.LocalDriver

	// 注册路由映射
	p.POST("/api/admin/upload/:resource/handle", "Handle")
	p.POST("/api/admin/upload/:resource/base64Handle", "HandleFromBase64")

	return p
}

// 执行上传
func (p *Template) Handle(ctx *builder.Context) interface{} {
	var (
		result *storage.FileInfo
		err    error
	)

	limitW := ctx.Query("limitW", "")
	limitH := ctx.Query("limitH", "")

	contentTypes := strings.Split(ctx.Header("Content-Type"), "; ")
	if len(contentTypes) != 2 {
		return msg.Error("Content-Type error", "")
	}
	if contentTypes[0] != "multipart/form-data" {
		return msg.Error("Content-Type must use multipart/form-data", "")
	}

	limitSize := reflect.
		ValueOf(ctx.Template).
		Elem().
		FieldByName("LimitSize").Int()

	limitType := reflect.
		ValueOf(ctx.Template).
		Elem().
		FieldByName("LimitType").Interface()

	limitImageWidth := int(reflect.
		ValueOf(ctx.Template).
		Elem().
		FieldByName("LimitImageWidth").Int())

	if limitW.(string) != "" {
		getLimitImageWidth, err := strconv.Atoi(limitW.(string))
		if err == nil {
			limitImageWidth = getLimitImageWidth
		}
	}

	limitImageHeight := int(reflect.
		ValueOf(ctx.Template).
		Elem().
		FieldByName("LimitImageHeight").Int())

	if limitH.(string) != "" {
		getLimitImageWidth, err := strconv.Atoi(limitH.(string))
		if err == nil {
			limitImageWidth = getLimitImageWidth
		}
	}

	driver := reflect.
		ValueOf(ctx.Template).
		Elem().
		FieldByName("Driver").String()

	ossConfig := reflect.
		ValueOf(ctx.Template).
		Elem().
		FieldByName("OSSConfig").Interface()

	savePath := reflect.
		ValueOf(ctx.Template).
		Elem().
		FieldByName("SavePath").String()

	byteReader := bytes.NewReader(ctx.Body())
	multipartReader := multipart.NewReader(byteReader, strings.TrimLeft(contentTypes[1], "boundary="))
	for p, err := multipartReader.NextPart(); err != io.EOF; p, err = multipartReader.NextPart() {
		if p.FormName() == "file" {
			fileData, _ := ioutil.ReadAll(p)
			fileSystem := storage.
				New(&storage.Config{
					LimitSize:        limitSize,
					LimitType:        limitType.([]string),
					LimitImageWidth:  limitImageWidth,
					LimitImageHeight: limitImageHeight,
					Driver:           driver,
					CheckFileExist:   true,
					OSSConfig:        ossConfig.(*storage.OSSConfig),
				}).
				Reader(&storage.File{
					Header:  p.Header,
					Name:    p.FileName(),
					Content: fileData,
				})

			// 上传前回调
			getFileSystem, fileInfo, err := ctx.Template.(interface {
				BeforeHandle(ctx *builder.Context, fileSystem *storage.FileSystem) (*storage.FileSystem, *storage.FileInfo, error)
			}).BeforeHandle(ctx, fileSystem)
			if err != nil {
				return msg.Error(err.Error(), "")
			}
			if fileInfo != nil {
				return ctx.Template.(interface {
					AfterHandle(ctx *builder.Context, result *storage.FileInfo) interface{}
				}).AfterHandle(ctx, fileInfo)
			}

			result, err = getFileSystem.
				WithImageWH().
				RandName().
				Path(savePath).
				Save()
		}
	}

	if err != nil {
		return msg.Error(err.Error(), "")
	}

	return ctx.Template.(interface {
		AfterHandle(ctx *builder.Context, result *storage.FileInfo) interface{}
	}).AfterHandle(ctx, result)
}

// 通过Base64执行上传
func (p *Template) HandleFromBase64(ctx *builder.Context) interface{} {
	var (
		result *storage.FileInfo
		err    error
	)

	limitW := ctx.Query("limitW", "")
	limitH := ctx.Query("limitH", "")

	data := map[string]interface{}{}
	if err := ctx.BodyParser(&data); err != nil {
		return msg.Error(err.Error(), "")
	}
	if data["file"] == nil {
		return msg.Error("参数错误", "")
	}

	files := strings.Split(data["file"].(string), ",")
	if len(files) != 2 {
		return msg.Error("格式错误", "")
	}

	fileData, err := base64.StdEncoding.DecodeString(files[1]) //成图片文件并把文件写入到buffer
	if err != nil {
		return msg.Error(err.Error(), "")
	}

	limitSize := reflect.
		ValueOf(ctx.Template).
		Elem().
		FieldByName("LimitSize").Int()

	limitType := reflect.
		ValueOf(ctx.Template).
		Elem().
		FieldByName("LimitType").Interface()

	limitImageWidth := int(reflect.
		ValueOf(ctx.Template).
		Elem().
		FieldByName("LimitImageWidth").Int())

	if limitW.(string) != "" {
		getLimitImageWidth, err := strconv.Atoi(limitW.(string))
		if err == nil {
			limitImageWidth = getLimitImageWidth
		}
	}

	limitImageHeight := int(reflect.
		ValueOf(ctx.Template).
		Elem().
		FieldByName("LimitImageHeight").Int())

	if limitH.(string) != "" {
		getLimitImageWidth, err := strconv.Atoi(limitH.(string))
		if err == nil {
			limitImageWidth = getLimitImageWidth
		}
	}

	savePath := reflect.
		ValueOf(ctx.Template).
		Elem().
		FieldByName("SavePath").String()

	driver := reflect.
		ValueOf(ctx.Template).
		Elem().
		FieldByName("Driver").String()

	ossConfig := reflect.
		ValueOf(ctx.Template).
		Elem().
		FieldByName("OSSConfig").Interface()

	fileSystem := storage.
		New(&storage.Config{
			LimitSize:        limitSize,
			LimitType:        limitType.([]string),
			LimitImageWidth:  limitImageWidth,
			LimitImageHeight: limitImageHeight,
			Driver:           driver,
			CheckFileExist:   true,
			OSSConfig:        ossConfig.(*storage.OSSConfig),
		}).
		Reader(&storage.File{
			Content: fileData,
		})

	// 上传前回调
	getFileSystem, fileInfo, err := ctx.Template.(interface {
		BeforeHandle(ctx *builder.Context, fileSystem *storage.FileSystem) (*storage.FileSystem, *storage.FileInfo, error)
	}).BeforeHandle(ctx, fileSystem)
	if err != nil {
		return msg.Error(err.Error(), "")
	}
	if fileInfo != nil {
		return ctx.Template.(interface {
			AfterHandle(ctx *builder.Context, result *storage.FileInfo) interface{}
		}).AfterHandle(ctx, fileInfo)
	}

	result, err = getFileSystem.
		WithImageWH().
		RandName().
		Path(savePath).
		Save()

	if err != nil {
		return msg.Error(err.Error(), "")
	}

	return ctx.Template.(interface {
		AfterHandle(ctx *builder.Context, result *storage.FileInfo) interface{}
	}).AfterHandle(ctx, result)
}

// 上传前回调
func (p *Template) BeforeHandle(ctx *builder.Context, fileSystem *storage.FileSystem) (*storage.FileSystem, *storage.FileInfo, error) {

	return fileSystem, nil, nil
}

// 执行上传
func (p *Template) AfterHandle(ctx *builder.Context, result *storage.FileInfo) interface{} {

	return msg.Success("上传成功", "", result)
}
