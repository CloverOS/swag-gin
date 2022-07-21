package swag

import (
	"github.com/Xuanwo/gg"
	"github.com/go-openapi/spec"
	"os"
	"path/filepath"
)

type router struct {
}

type RouteInfos struct {
	Method     string `json:"method"`      // method
	Path       string `json:"path"`        //path
	HandlerFun string `json:"handler_fun"` //handlerFun
	Summary    string `json:"summary"`     //Summary
	RouteGroup
}

type RouteGroup struct {
	GroupName string `json:"name"`
}

type Routes struct {
	FilePath string
	PkgName  string
}

type GenConfig struct {
	GinServerPackage string
	GinRouterPath    string
}

var GinRouter = new(router)

func (*router) RegisterRouter(p *Parser, g GenConfig) error {
	swagger := p.GetSwagger()
	basePath := swagger.BasePath
	routes := make(map[Routes][]RouteInfos)
	for path, v := range swagger.SwaggerProps.Paths.Paths {
		method, operation := getRoute(v)
		groupName := "unknown"
		if len(operation.Tags) > 0 {
			groupName = operation.Tags[0]
		}
		route := RouteInfos{
			Method:     method,
			Path:       basePath + path,
			HandlerFun: p.HandlerFunc[path],
			Summary:    operation.Summary,
			RouteGroup: RouteGroup{
				GroupName: groupName,
			},
		}
		routes[Routes{
			FilePath: p.FilePathHandlerFunc[path],
			PkgName:  p.PkgName[path],
		}] = append(routes[Routes{
			FilePath: p.FilePathHandlerFunc[path],
			PkgName:  p.PkgName[path],
		}], route)
	}
	return createFile(routes, g, p)
}

func createFile(routes map[Routes][]RouteInfos, config GenConfig, p *Parser) error {
	for filePath, infos := range routes {
		g := gg.New()
		f := g.NewGroup()
		f.AddPackage(filePath.PkgName)
		f.NewImport().
			AddPath("github.com/gin-gonic/gin")
		functions := f.NewFunction("Register").AddParameter("r", "*gin.RouterGroup")
		for _, v := range infos {
			if v.Method == "get" {
				functions.AddBody(gg.String(`r.GET(%s,%s)`, "\""+v.Path+"\"", v.HandlerFun))
			}
			if v.Method == "post" {
				functions.AddBody(gg.String(`r.POST(%s,%s)`, "\""+v.Path+"\"", v.HandlerFun))
			}
			if v.Method == "head" {
				functions.AddBody(gg.String(`r.HEAD(%s,%s)`, "\""+v.Path+"\"", v.HandlerFun))
			}
			if v.Method == "put" {
				functions.AddBody(gg.String(`r.PUT(%s,%s)`, "\""+v.Path+"\"", v.HandlerFun))
			}
			if v.Method == "delete" {
				functions.AddBody(gg.String(`r.DELETE(%s,%s)`, "\""+v.Path+"\"", v.HandlerFun))
			}
			if v.Method == "options" {
				functions.AddBody(gg.String(`r.OPTIONS(%s,%s)`, "\""+v.Path+"\"", v.HandlerFun))
			}
			if v.Method == "patch" {
				functions.AddBody(gg.String(`r.PATCH(%s,%s)`, "\""+v.Path+"\"", v.HandlerFun))
			}
		}
		exists, err := pathExists(filePath.FilePath + string(filepath.Separator) + "router.go")
		if err != nil {
			return err
		}
		if !exists {
			err = g.WriteFile(filePath.FilePath + string(filepath.Separator) + "router.go")
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
