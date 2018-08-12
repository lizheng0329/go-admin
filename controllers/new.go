package controller

import (
	"bytes"
	"github.com/valyala/fasthttp"
	"goAdmin/auth"
	"goAdmin/menu"
	"goAdmin/models"
	"goAdmin/modules/file"
	"goAdmin/template"
	"strings"
)

// 显示新建表单
func ShowNewForm(ctx *fasthttp.RequestCtx) {
	defer GlobalDeferHandler(ctx)

	prefix := ctx.UserValue("prefix").(string)

	buffer := new(bytes.Buffer)

	id := string(ctx.QueryArgs().Peek("id")[:])

	page := string(ctx.QueryArgs().Peek("page")[:])
	if page == "" {
		page = "1"
	}
	pageSize := string(ctx.QueryArgs().Peek("pageSize")[:])
	if pageSize == "" {
		pageSize = "10"
	}

	url := "/new/" + prefix + "?id=" + id
	previous := "/info/" + prefix + "?page=" + page + "&pageSize=" + pageSize

	token := auth.TokenHelper.AddToken()

	if string(ctx.Request.Header.Peek("X-PJAX")[:]) == "true" {
		template.NewPanelPjax(models.GlobalTableList[prefix].Form.FormList, url, previous, id,
			models.GlobalTableList[prefix].Form.Title, models.GlobalTableList[prefix].Form.Description,
			models.ErrStruct{"hidden", ""}, token, buffer)
	} else {
		template.NewPanel(models.GlobalTableList[prefix].Form.FormList, url, previous, id,
			(*menu.GlobalMenu).GlobalMenuList, models.GlobalTableList[prefix].Form.Title,
			models.GlobalTableList[prefix].Form.Description, token, buffer)
	}

	ctx.Response.AppendBody(buffer.Bytes())
	ctx.Response.Header.Add("Content-Type", "text/html; charset=utf-8")
}

// 新建数据
func NewForm(ctx *fasthttp.RequestCtx) {

	defer GlobalDeferHandler(ctx)

	token := string(ctx.FormValue("_t"))

	if !auth.TokenHelper.CheckToken(token) {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.WriteString(`{"code":400, "msg":"编辑失败"}`)
		return
	}

	prefix := ctx.UserValue("prefix").(string)

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.SetStatusCode(500)
		return
	}

	// 处理上传文件，目前仅仅支持传本地
	if len((*form).File) > 0 {
		file.GetFileEngine("local").Upload(form)
	}

	if prefix == "manager" { // 管理员管理新建
		NewManager((*form).Value)
	} else if prefix == "roles" { // 管理员角色管理新建
		NewRole((*form).Value)
	} else {
		models.GlobalTableList[prefix].InsertDataFromDatabase(prefix, (*form).Value)
	}

	models.RefreshGlobalTableList()

	previous := string(ctx.FormValue("_previous_"))
	prevUrlArr := strings.Split(previous, "?")
	paramArr := strings.Split(prevUrlArr[1], "&")
	page := "1"
	pageSize := "10"
	sort := "id"
	sort_type := "desc"

	for i := 0; i < len(paramArr); i++ {
		if strings.Index(paramArr[i], "pageSize") >= 0 {
			pageSize = strings.Split(paramArr[i], "=")[1]
		} else {
			if strings.Index(paramArr[i], "page") >= 0 {
				page = strings.Split(paramArr[i], "=")[1]
			} else if strings.Index(paramArr[i], "sort") >= 0 {
				sort = strings.Split(paramArr[i], "=")[1]
			} else {
				sort_type = strings.Split(paramArr[i], "=")[1]
			}
		}
	}

	thead, infoList, paginator, title, description := models.GlobalTableList[prefix].GetDataFromDatabase(map[string]string{
		"page":      page,
		"path":      prevUrlArr[0],
		"sortField": sort,
		"sortType":  sort_type,
		"prefix":    prefix,
		"pageSize":  pageSize,
	})

	menu.GlobalMenu.SetActiveClass(previous)
	paginator["success"] = true

	buffer := new(bytes.Buffer)

	template.InfoListPjax(infoList, (*menu.GlobalMenu).GlobalMenuList, thead, paginator, title, description, buffer)

	ctx.Response.AppendBody(buffer.Bytes())
	ctx.Response.Header.Add("Content-Type", "text/html; charset=utf-8")
	ctx.Response.Header.Add("X-PJAX-URL", previous)
}
