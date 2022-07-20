package api

import (
	"net/http"

	. "github.com/swaggo/swag/testdata/simple/cross"
	_ "github.com/swaggo/swag/testdata/simple/web"
)

// @Summary Add a new pet to the store
// @Description get string by ID
// @ID get-string-by-int
// @Tags testmodel
// @Accept  json
// @Produce  json
// @Param   some_id      path   int     true  "Some ID" Format(int64)
// @Param   some_id      body web.Pet true  "Some ID"
// @Success 200 {string} string	"ok"
// @Failure 400 {object} web.APIError "We need ID!!"
// @Failure 404 {object} web.APIError "Can not find ID"
// @Router /testapi/get-string-by-int/{some_id} [get]
func GetStringByInt(w http.ResponseWriter, r *http.Request) {
	_ = Cross{}
	//write your code
}
