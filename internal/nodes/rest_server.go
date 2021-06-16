package nodes

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/iwind/TeaGo/maps"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"regexp"
	"time"
)

var servicePathReg = regexp.MustCompile(`^/([a-zA-Z0-9]+)/([a-zA-Z0-9]+)$`)
var servicesMap = map[string]reflect.Value{
	"APIAccessTokenService": reflect.ValueOf(new(services.APIAccessTokenService)),
	"HTTPAccessLogService":  reflect.ValueOf(new(services.HTTPAccessLogService)),
	"IPItemService":         reflect.ValueOf(new(services.IPItemService)),
}

type RestServer struct{}

func (this *RestServer) Listen(listener net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", this.handle)
	server := &http.Server{}
	server.Handler = mux
	return server.Serve(listener)
}

func (this *RestServer) ListenHTTPS(listener net.Listener, tlsConfig *tls.Config) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", this.handle)
	server := &http.Server{}
	server.Handler = mux
	server.TLSConfig = tlsConfig
	return server.ServeTLS(listener, "", "")
}

func (this *RestServer) handle(writer http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	// 是否显示Pretty后的JSON
	shouldPretty := req.Header.Get("Edge-Response-Pretty") == "on"

	// 欢迎页
	if path == "/" {
		this.writeJSON(writer, maps.Map{
			"code":    200,
			"message": "Welcome to API",
			"data":    maps.Map{},
		}, shouldPretty)
		return
	}

	matches := servicePathReg.FindStringSubmatch(path)
	if len(matches) != 3 {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	serviceName := matches[1]
	methodName := matches[2]

	serviceType, ok := servicesMap[serviceName]
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	method := serviceType.MethodByName(methodName)
	if !method.IsValid() {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	if method.Type().NumIn() != 2 || method.Type().NumOut() != 2 {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	if method.Type().In(0).Name() != "Context" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	// 上下文
	ctx := context.Background()

	if serviceName != "APIAccessTokenService" || methodName != "GetAPIAccessToken" {
		// 校验TOKEN
		token := req.Header.Get("Edge-Access-Token")
		if len(token) == 0 {
			this.writeJSON(writer, maps.Map{
				"code":    400,
				"data":    maps.Map{},
				"message": "require 'Edge-Access-Token' header",
			}, shouldPretty)
			return
		}

		accessToken, err := models.SharedAPIAccessTokenDAO.FindAccessToken(nil, token)
		if err != nil {
			this.writeJSON(writer, maps.Map{
				"code":    400,
				"data":    maps.Map{},
				"message": "server error: " + err.Error(),
			}, shouldPretty)
			return
		}

		if accessToken == nil || int64(accessToken.ExpiredAt) < time.Now().Unix() {
			this.writeJSON(writer, maps.Map{
				"code":    400,
				"data":    maps.Map{},
				"message": "invalid access token",
			}, shouldPretty)
			return
		}

		if accessToken.UserId > 0 {
			ctx = rpcutils.NewPlainContext("user", int64(accessToken.UserId))
		} else {
			// TODO 支持更多类型的角色
			this.writeJSON(writer, maps.Map{
				"code":    400,
				"data":    maps.Map{},
				"message": "not supported role",
			}, shouldPretty)
			return
		}
	}

	// TODO 需要防止BODY过大攻击
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(err.Error()))
		return
	}

	// 请求数据
	reqValue := reflect.New(method.Type().In(1).Elem()).Interface()
	err = json.Unmarshal(body, reqValue)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte("Decode request failed: " + err.Error() + ". Request body should be a valid JSON data"))
		return
	}

	result := method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(reqValue)})
	resultErr := result[1].Interface()
	if resultErr != nil {
		e, ok := resultErr.(error)
		if ok {
			this.writeJSON(writer, maps.Map{
				"code":    400,
				"message": e.Error(),
				"data":    maps.Map{},
			}, shouldPretty)
		} else {
			this.writeJSON(writer, maps.Map{
				"code":    500,
				"message": "server error: server should return a error object, but return a " + result[1].Type().String(),
				"data":    maps.Map{},
			}, shouldPretty)
		}
	} else { // 没有返回错误
		data := maps.Map{
			"code":    200,
			"message": "ok",
			"data":    result[0].Interface(),
		}
		var dataJSON []byte
		if shouldPretty {
			dataJSON = data.AsPrettyJSON()
		} else {
			dataJSON = data.AsJSON()
		}
		if err != nil {
			this.writeJSON(writer, maps.Map{
				"code":    500,
				"message": "server error: marshal json failed: " + err.Error(),
				"data":    maps.Map{},
			}, shouldPretty)
		} else {
			writer.Header().Set("Content-Type", "application/json; charset=utf-8")

			_, _ = writer.Write(dataJSON)
		}
	}
}

func (this *RestServer) writeJSON(writer http.ResponseWriter, v maps.Map, pretty bool) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")

	if pretty {
		_, _ = writer.Write(v.AsPrettyJSON())
	} else {
		_, _ = writer.Write(v.AsJSON())
	}
}
