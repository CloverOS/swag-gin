package swag

import (
	"errors"
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/go-openapi/spec"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type router struct {
}

type RouteInfos struct {
	Method     string `json:"method"`      // method
	Path       string `json:"path"`        //path
	HandlerFun string `json:"handler_fun"` //handlerFun
	Summary    string `json:"summary"`     //Summary
	Public     bool   `json:"public"`      //is public router
	RouteGroup
}

type RouteGroup struct {
	GroupName string `json:"name"`
}

type Routes struct {
	FilePath        string
	PkgName         string
	RouteModuleName string
}

type GenConfig struct {
	AutoCover bool   //自动覆盖
	OutputDir string //输出文件夹
	OutputPkg string //输出包名
}

// IntRouterPath
// @Description: 路由注册方法
type IntRouterPath struct {
	FilePath        string
	PkgName         string
	RouteModuleName string
}

// MyRouterPath 路由注册函数
type MyRouterPath struct {
	Method      string   `json:"method"`     //路由方法
	Path        string   `json:"path"`       //路由路径
	Fun         string   `json:"fun"`        //路由函数
	Remarks     string   `json:"remarks"`    //备注(用于列表处理时显示)
	Auth        bool     `json:"-"`          //需要授权
	Allow       string   `json:"-"`          //访问限制
	AuthHandler []string `json:"-"`          //中间件
	MenuItems   string   `json:"menu_items"` // 菜单分类
}

var GinRouter = new(router)

func (*router) RegisterRouterPath(p *Parser, g GenConfig) error {
	swagger := p.GetSwagger()
	route := make(map[IntRouterPath][]MyRouterPath)
	for path, v := range swagger.SwaggerProps.Paths.Paths {
		method, operation := getRoute(v)
		var handler []string
		if len(operation.Security) > 0 {
			a := operation.Security[0]
			for s, i := range a {
				switch s {
				case "auth":
					handler = []string{"auth", "JwtAuthMiddleware()", "handler", "AuthCheckRole(util.GetE())"}
				case "wx":
					handler = []string{"auth", "WeiXinAuth()"}
				default:
					if len(i) > 0 {
						handler = i
					}
				}
			}
		}
		r := MyRouterPath{
			Method:      method,
			Path:        path,
			Fun:         p.HandlerFunc[path],
			Auth:        len(handler) >= 2,
			AuthHandler: handler,
			Remarks:     operation.Summary,
			MenuItems:   operation.Tags[0],
		}
		route[IntRouterPath{
			FilePath:        p.FilePathHandlerFunc[path],
			PkgName:         p.PkgName[path],
			RouteModuleName: p.HandlerFuncModules[path],
		}] = append(route[IntRouterPath{
			FilePath:        p.FilePathHandlerFunc[path],
			PkgName:         p.PkgName[path],
			RouteModuleName: p.HandlerFuncModules[path],
		}], r)
	}
	err := genDocFilePath(route, g)
	if err != nil {
		return err
	}
	return nil
}

func genDocFilePath(routes map[IntRouterPath][]MyRouterPath, config GenConfig) error {
	finalPath := strings.ReplaceAll(config.OutputDir, "/docs", "/"+config.OutputPkg)
	exists, err := pathExists(finalPath)
	if err != nil {
		return err
	}
	if !exists {
		err := os.MkdirAll(finalPath, 0755)
		if err != nil {
			return err
		}
	}
	f := jen.NewFilePath(finalPath)
	f.ImportName("github.com/gin-gonic/gin", "")
	f.Type().Id("IntRouterPath").Struct(
		jen.Id("FilePath").String().Tag(map[string]string{"json": "file_path"}),
		jen.Id("PkgName").String().Tag(map[string]string{"json": "pkg_name"}),
		jen.Id("RouteModuleName").String().Tag(map[string]string{"json": "route_module_name"}),
	)
	f.Line()
	f.Type().Id("MyRouterPath").Struct(
		jen.Id("Method").String().Tag(map[string]string{"json": "method"}).Comment("请求方法"),
		jen.Id("Path").String().Tag(map[string]string{"json": "path"}).Comment(" 路由路径"),
		jen.Id("Fun").UnionFunc(func(group *jen.Group) {
			group.Id("gin.HandlerFunc")
		}).Tag(map[string]string{"json": "-"}).Comment(" 路由方法"),
		jen.Id("Remarks").String().Tag(map[string]string{"json": "remarks"}).Comment(" 备注"),
		jen.Id("Auth").Bool().Tag(map[string]string{"json": "auth"}).Comment(" 是否鉴权"),
		jen.Id("AuthHandler").UnionFunc(func(group *jen.Group) {
			group.Id("[]gin.HandlerFunc")
		}).Tag(map[string]string{"json": "-"}).Comment(" 鉴权中间件"),
		jen.Id("MenuItems").String().Tag(map[string]string{"json": "menu_items"}).Comment(" 菜单分类"),
	)
	var values []jen.Code
	for filePath, infos := range routes {
		for _, info := range infos {
			packageName, err := GetPackageName(filePath.FilePath)
			split := strings.Split(info.Fun, ".")
			var handlerFuncName = info.Fun
			prefix := ""
			if len(split) > 1 {
				prefix = split[0]
				packageName = filepath.Join(packageName, prefix)
				var temp string
				handlerFuncName = strings.Replace(info.Fun, prefix+".", temp, 1)
			}
			packageName = strings.Replace(packageName, "\\", "/", -1)
			f.ImportName(packageName, prefix)
			if err != nil {
				return err
			}
			if len(info.AuthHandler) > 0 {
				if !isEven(len(info.AuthHandler)) {
					return errors.New("鉴权设置异常，请检查Security字段")
				}
				f.ImportName("ShuHeSdk/util", "")
				values = append(values, jen.Values(jen.Dict{
					jen.Id("Method"):  jen.Lit(strings.ToTitle(info.Method)),
					jen.Id("Path"):    jen.Lit(info.Path),
					jen.Id("Fun"):     jen.Qual(packageName, handlerFuncName),
					jen.Id("Remarks"): jen.Lit(info.Remarks),
					jen.Id("Auth"):    jen.Lit(info.Auth),
					jen.Id("AuthHandler"): jen.ListFunc(func(group *jen.Group) {
						s := group.Id("[]gin.HandlerFunc")
						var cs []jen.Code
						for i, _ := range info.AuthHandler {
							if isEven(i) || i == 0 {
								cs = append(cs, jen.Qual(info.AuthHandler[i], info.AuthHandler[i+1]))
							}
						}
						s.Values(cs...)
					}),
					jen.Id("MenuItems"): jen.Lit(info.MenuItems),
				}))
			} else {
				values = append(values, jen.Values(jen.Dict{
					jen.Id("Method"):    jen.Lit(strings.ToTitle(info.Method)),
					jen.Id("Path"):      jen.Lit(info.Path),
					jen.Id("Fun"):       jen.Qual(packageName, handlerFuncName),
					jen.Id("Remarks"):   jen.Lit(info.Remarks),
					jen.Id("Auth"):      jen.Lit(info.Auth),
					jen.Id("MenuItems"): jen.Lit(info.MenuItems),
				}))
			}
		}
	}
	f.Func().Id("IntRouterPaths").Params().Index().Id("MyRouterPath").Block(
		jen.Return(jen.Index().Id("MyRouterPath").Values(values...)))
	resource := filepath.Join(finalPath, "IntRouterPaths.go")
	exists, err = pathExists(resource)
	if err != nil {
		return err
	}
	if !exists || config.AutoCover {
		file, err := os.Create(resource)
		if err != nil {
			return err
		}
		err = f.Render(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func (*router) RegisterRouter(p *Parser, g GenConfig) error {
	swagger := p.GetSwagger()
	routes := make(map[Routes][]RouteInfos)
	for path, v := range swagger.SwaggerProps.Paths.Paths {
		method, operation := getRoute(v)
		groupName := "unknown"
		if len(operation.Tags) > 0 {
			groupName = operation.Tags[0]
		}
		route := RouteInfos{
			Method:     method,
			Path:       path,
			HandlerFun: p.HandlerFunc[path],
			Summary:    operation.Summary,
			RouteGroup: RouteGroup{
				GroupName: groupName,
			},
			Public: len(operation.Security) < 1,
		}
		routes[Routes{
			FilePath:        p.FilePathHandlerFunc[path],
			PkgName:         p.PkgName[path],
			RouteModuleName: p.HandlerFuncModules[path],
		}] = append(routes[Routes{
			FilePath:        p.FilePathHandlerFunc[path],
			PkgName:         p.PkgName[path],
			RouteModuleName: p.HandlerFuncModules[path],
		}], route)
	}
	err := genDocFile(routes, g)
	if err != nil {
		return err
	}
	return genGoFile(routes, g)
}

func genDocFile(routes map[Routes][]RouteInfos, config GenConfig) error {
	f := jen.NewFilePath(config.OutputDir)
	f.Type().Id("RouteGroup").Struct(
		jen.Id("GroupName").String().Tag(map[string]string{"json": "name"}),
	)
	f.Line()
	f.Type().Id("RouteInfos").Struct(
		jen.Id("Method").String().Tag(map[string]string{"json": "method"}).Comment(" method"),
		jen.Id("Path").String().Tag(map[string]string{"json": "path"}).Comment(" path"),
		jen.Id("HandlerFun").String().Tag(map[string]string{"json": "handler_fun"}).Comment(" handlerFun"),
		jen.Id("Summary").String().Tag(map[string]string{"json": "summary"}).Comment(" Summary"),
		jen.Id("Public").Bool().Tag(map[string]string{"json": "public"}).Comment(" is public router"),
		jen.Id("RouteGroup"),
	)
	var values []jen.Code
	for _, infos := range routes {
		for _, info := range infos {
			values = append(values, jen.Values(jen.Dict{
				jen.Id("Method"):     jen.Lit(info.Method),
				jen.Id("Path"):       jen.Lit(info.Path),
				jen.Id("HandlerFun"): jen.Lit(info.HandlerFun),
				jen.Id("Summary"):    jen.Lit(info.Summary),
				jen.Id("Public"):     jen.Lit(info.Public),
				jen.Id("RouteGroup"): jen.Id("RouteGroup").Values(jen.Dict{
					jen.Id("GroupName"): jen.Lit(info.GroupName),
				}),
			}))
		}
	}
	f.Func().Id("GetRouteInfos").Params().Index().Id("RouteInfos").Block(
		jen.Return(jen.Index().Id("RouteInfos").Values(values...)))
	resource := filepath.Join(config.OutputDir, "resource.go")
	exists, err := pathExists(resource)
	if err != nil {
		return err
	}
	if !exists || config.AutoCover {
		file, err := os.Create(resource)
		if err != nil {
			return err
		}
		err = f.Render(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func genGoFile(routes map[Routes][]RouteInfos, config GenConfig) error {
	for filePath, infos := range routes {
		finalPath := filepath.Join(filePath.FilePath, "router.go")
		f := jen.NewFilePath(filePath.PkgName)
		f.ImportName("github.com/gin-gonic/gin", "")
		rParams := jen.Id("r").Op("*").Id("gin").Dot("RouterGroup")
		var publicCode []jen.Code
		var privateCode []jen.Code
		for _, v := range infos {
			packageName, err := GetPackageName(filePath.FilePath)
			split := strings.Split(v.HandlerFun, ".")
			var handlerFuncName = v.HandlerFun
			prefix := ""
			if len(split) > 1 {
				prefix = split[0]
				packageName = filepath.Join(packageName, prefix)
				var temp string
				handlerFuncName = strings.Replace(v.HandlerFun, prefix+".", temp, 1)
			}
			f.ImportName(packageName, prefix)
			if err != nil {
				return err
			}
			if v.Public {
				if v.Method == "get" {
					publicCode = append(publicCode, jen.Id("r").Dot("GET").Call(jen.Id("\""+v.Path+"\""), jen.Qual(packageName, handlerFuncName)))
				}
				if v.Method == "post" {
					publicCode = append(publicCode, jen.Id("r").Dot("POST").Call(jen.Id("\""+v.Path+"\""), jen.Qual(packageName, handlerFuncName)))
				}
				if v.Method == "head" {
					publicCode = append(publicCode, jen.Id("r").Dot("HEAD").Call(jen.Id("\""+v.Path+"\""), jen.Qual(packageName, handlerFuncName)))
				}
				if v.Method == "put" {
					publicCode = append(publicCode, jen.Id("r").Dot("PUT").Call(jen.Id("\""+v.Path+"\""), jen.Qual(packageName, handlerFuncName)))
				}
				if v.Method == "delete" {
					publicCode = append(publicCode, jen.Id("r").Dot("DELETE").Call(jen.Id("\""+v.Path+"\""), jen.Qual(packageName, handlerFuncName)))
				}
				if v.Method == "options" {
					publicCode = append(publicCode, jen.Id("r").Dot("OPTIONS").Call(jen.Id("\""+v.Path+"\""), jen.Qual(packageName, handlerFuncName)))
				}
				if v.Method == "patch" {
					publicCode = append(publicCode, jen.Id("r").Dot("PATCH").Call(jen.Id("\""+v.Path+"\""), jen.Qual(packageName, handlerFuncName)))
				}
			} else {
				if v.Method == "get" {
					privateCode = append(privateCode, jen.Id("r").Dot("GET").Call(jen.Id("\""+v.Path+"\""), jen.Qual(packageName, handlerFuncName)))
				}
				if v.Method == "post" {
					privateCode = append(privateCode, jen.Id("r").Dot("POST").Call(jen.Id("\""+v.Path+"\""), jen.Qual(packageName, handlerFuncName)))
				}
				if v.Method == "head" {
					privateCode = append(privateCode, jen.Id("r").Dot("HEAD").Call(jen.Id("\""+v.Path+"\""), jen.Qual(packageName, handlerFuncName)))
				}
				if v.Method == "put" {
					privateCode = append(privateCode, jen.Id("r").Dot("PUT").Call(jen.Id("\""+v.Path+"\""), jen.Qual(packageName, handlerFuncName)))
				}
				if v.Method == "delete" {
					privateCode = append(privateCode, jen.Id("r").Dot("DELETE").Call(jen.Id("\""+v.Path+"\""), jen.Qual(packageName, handlerFuncName)))
				}
				if v.Method == "options" {
					privateCode = append(privateCode, jen.Id("r").Dot("OPTIONS").Call(jen.Id("\""+v.Path+"\""), jen.Qual(packageName, handlerFuncName)))
				}
				if v.Method == "patch" {
					privateCode = append(privateCode, jen.Id("r").Dot("PATCH").Call(jen.Id("\""+v.Path+"\""), jen.Qual(packageName, handlerFuncName)))
				}
			}
		}
		f.Func().Id("InitPublicRouter").Params(rParams).Block(publicCode...)
		f.Line()
		f.Func().Id("InitPrivateRouter").Params(rParams).Block(privateCode...)
		exists, err := pathExists(finalPath)
		if err != nil {
			return err
		}
		if !exists || config.AutoCover {
			file, err := os.Create(finalPath)
			if err != nil {
				return err
			}
			err = f.Render(file)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getRoute(pathItem spec.PathItem) (string, *spec.Operation) {
	if pathItem.Get != nil {
		return "get", pathItem.Get
	}
	if pathItem.Post != nil {
		return "post", pathItem.Post
	}
	if pathItem.Put != nil {
		return "put", pathItem.Put
	}
	if pathItem.Delete != nil {
		return "delete", pathItem.Delete
	}
	if pathItem.Head != nil {
		return "head", pathItem.Head
	}
	if pathItem.Options != nil {
		return "options", pathItem.Options
	}
	if pathItem.Patch != nil {
		return "patch", pathItem.Patch
	}
	return "unknown", nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetPackageName(searchDir string) (string, error) {
	cmd := exec.Command("go", "list", "-f={{.ImportPath}}")
	cmd.Dir = searchDir

	var stdout, stderr strings.Builder

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("execute go list command, %s, stdout:%s, stderr:%s", err, stdout.String(), stderr.String())
	}

	outStr, _ := stdout.String(), stderr.String()

	if outStr[0] == '_' { // will shown like _/{GOPATH}/src/{YOUR_PACKAGE} when NOT enable GO MODULE.
		outStr = strings.TrimPrefix(outStr, "_"+build.Default.GOPATH+"/src/")
	}

	f := strings.Split(outStr, "\n")

	outStr = f[0]

	return outStr, nil
}

func isEven(num int) bool {
	if num%2 == 0 {
		return true
	}
	return false
}
