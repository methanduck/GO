package ProxySVR

import (
	"fmt"
	httpRouter "github.com/julienschmidt/httprouter"
	"net/http"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
type (
	Post_Unavailable   struct{}
	Delete_Unavailable struct{}
)

type Resource interface {
	Uri() string
	Get(writer http.ResponseWriter, req *http.Request, ps httpRouter.Params) Response
	Put(writer http.ResponseWriter, req *http.Request, ps httpRouter.Params) Response
	Post(writer http.ResponseWriter, req *http.Request, ps httpRouter.Params) Response
	Delete(writer http.ResponseWriter, req *http.Request, ps httpRouter.Params) Response
}

func (Post_Unavailable) Post() Response {
	return Response{405, "", nil}
}
func (Delete_Unavailable) Delete() Response {
	return Response{405, "", nil}
}

func HTTP_Response(writer http.ResponseWriter, req *http.Request, ps httpRouter.Params) {

}

func Add_Resource(router *httpRouter.Router, resource Resource) {

	router.GET(resource.Uri(), func(writer http.ResponseWriter, request *http.Request, params httpRouter.Params) {
		res := resource.Get(writer, request, params)
		HTTP_Response()
	})
}
