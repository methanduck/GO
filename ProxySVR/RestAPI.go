package ProxySVR

import (
	"encoding/json"
	httpRouter "github.com/julienschmidt/httprouter"
	"net/http"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
type (
	Unavailable_Get    struct{}
	Unavailable_Put    struct{}
	Unavailable_Post   struct{}
	Unavailable_Delete struct{}
)

func (Unavailable_Post) Post() Response {
	return Response{http.StatusMethodNotAllowed, "", nil}
}
func (Unavailable_Put) Put() Response {
	return Response{http.StatusMethodNotAllowed, "", nil}
}
func (Unavailable_Delete) Delete() Response {
	return Response{http.StatusMethodNotAllowed, "", nil}
}
func (Unavailable_Get) Get() Response {
	return Response{http.StatusMethodNotAllowed, "", nil}
}

type Resource interface {
	Uri() string
	GET(writer http.ResponseWriter, req *http.Request, ps httpRouter.Params) Response
	PUT(writer http.ResponseWriter, req *http.Request, ps httpRouter.Params) Response
	POST(writer http.ResponseWriter, req *http.Request, ps httpRouter.Params) Response
	DELETE(writer http.ResponseWriter, req *http.Request, ps httpRouter.Params) Response
}

func HTTP_Response(writer http.ResponseWriter, req *http.Request, rsc Response) {
	data, err := json.Marshal(rsc)
	if err != nil {
		writer.WriteHeader(500)
	}
	writer.WriteHeader(rsc.Code)
	writer.Write(data)
}

func Add_Resource(router *httpRouter.Router, rsc Resource) {
	router.GET(rsc.Uri(), func(writer http.ResponseWriter, request *http.Request, params httpRouter.Params) {
		res := rsc.GET(writer, request, params)
		HTTP_Response(writer, request, res)
	})

	router.POST(rsc.Uri(), func(writer http.ResponseWriter, request *http.Request, params httpRouter.Params) {
		res := rsc.POST(writer, request, params)
		HTTP_Response(writer, request, res)
	})

	router.PUT(rsc.Uri(), func(writer http.ResponseWriter, request *http.Request, params httpRouter.Params) {
		res := rsc.PUT(writer, request, params)
		HTTP_Response(writer, request, res)
	})

	router.DELETE(rsc.Uri(), func(writer http.ResponseWriter, request *http.Request, params httpRouter.Params) {
		res := rsc.DELETE(writer, request, params)
		HTTP_Response(writer, request, res)
	})
}
