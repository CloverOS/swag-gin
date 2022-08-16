package api

import (
	"github.com/gin-gonic/gin"
	. "github.com/swaggo/swag/testdata/simple/cross"
	_ "github.com/swaggo/swag/testdata/simple/web"
	"gorm.io/gorm"
)

// GetStringByInt
// @Summary Add a new pet to the store
// @Description get string by ID
// @Tags testmodel
// @Accept  json
// @Produce  json
// @Param   some_id      path   int     true  "Some ID" Format(int64)
// @Param   some_id      body web.Pet true  "Some ID"
// @Success 200 {string} string	"ok"
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Security    AccessToken
// @Router /testapi/get-string-by-int/{some_id} [get]
func GetStringByInt(c *gin.Context) {
	_ = Cross{}
	//write your code
}

// GetStringByInt2
// @Summary Add a new pet to the store
// @Description get string by ID
// @Tags testmodel
// @Accept  json
// @Produce  json
// @Param   some_id      path   int     true  "Some ID" Format(int64)
// @Param   some_id      body web.Pet true  "Some ID"
// @Success 200 {string} string	"ok"
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Router /testapi/get-string-by-int2/{some_id} [get]
func GetStringByInt2(c *gin.Context) {
	_ = Cross{}
	//write your code
}

// GetStringByInt3
// @Summary Add a new pet to the store
// @Description get string by ID
// @Tags testmodel
// @Accept  json
// @Produce  json
// @Param   some_id      path   int     true  "Some ID" Format(int64)
// @Param   some_id      body web.Pet true  "Some ID"
// @Success 200 {string} string	"ok"
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Security    AccessToken
// @Router /testapi/get-string-by-int3/{some_id} [post]
func GetStringByInt3(c *gin.Context) {
	_ = Cross{}
	//write your code
}

type TestModel struct {
	gorm.Model `swaggerignore:"true"`
	T1         Test1 `json:"t_1"`
	T2         T2    `json:"t_2"`
}
type Test1 struct {
	T1 string `json:"t_1"`
}
type T2 struct {
	T2 string `json:"t_2"`
}

// GetStringByInt4
// @Summary Add a new pet to the store
// @Description get string by ID
// @Tags testmodel
// @Accept  json
// @Produce  json
// @Param _ body TestModel true " "
// @Success 200 {string} string	"ok"
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Security    AccessToken[ShuHeSdk/pkg/auth,JwtAuthMiddleware(),ShuHeSdk/pkg/handler,AuthCheckRole(util.GetE())]
// @Router /testapi/get-string-by-int4/{some_id} [post]
func GetStringByInt4(c *gin.Context) {
	_ = Cross{}
}
