package swag

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/go-openapi/spec"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type router struct {
}

type RouteInfos struct {
	Method     string `json:"method"`      // method
	Path       string `json:"path"`        //path
	BasePath   string `json:"base_path"`   //basePath
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
}

var GinRouter = new(router)

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
			BasePath:   swagger.BasePath,
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
		jen.Id("BasePath").String().Tag(map[string]string{"json": "base_path"}).Comment(" BasePath"),
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
				jen.Id("BasePath"):   jen.Lit(info.BasePath),
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
		f.ImportName("github.com/gin-gonic/gin", "gin")
		rParams := jen.Id("r").Op("*").Qual("github.com/gin-gonic/gin", "RouterGroup")
		var publicCode []jen.Code
		var privateCode []jen.Code
		var sortPath []string
		tempInfos := make(map[string]RouteInfos)
		for _, v := range infos {
			sortPath = append(sortPath, v.Path)
			tempInfos[v.Path] = v
		}
		sort.Strings(sortPath)
		i := 0
		for range tempInfos {
			v := tempInfos[sortPath[i]]
			i++
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
			packageName = strings.Replace(packageName, "\\", "/", -1)
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
