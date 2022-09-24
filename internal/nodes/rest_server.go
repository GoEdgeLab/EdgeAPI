package nodes

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/sizes"
	"github.com/iwind/TeaGo/maps"
	"io"
	"net"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var servicePathReg = regexp.MustCompile(`^/([a-zA-Z0-9]+)/([a-zA-Z0-9]+)$`)
var restServicesMap = map[string]reflect.Value{
	"APIAccessTokenService": reflect.ValueOf(new(services.APIAccessTokenService)),
}

type RestServer struct{}

func (this *RestServer) Listen(listener net.Listener) error {
	var mux = http.NewServeMux()
	mux.HandleFunc("/", this.handle)
	var server = &http.Server{}
	server.Handler = mux
	return server.Serve(listener)
}

func (this *RestServer) ListenHTTPS(listener net.Listener, tlsConfig *tls.Config) error {
	var mux = http.NewServeMux()
	mux.HandleFunc("/", this.handle)
	server := &http.Server{}
	server.Handler = mux
	server.TLSConfig = tlsConfig
	return server.ServeTLS(listener, "", "")
}

func (this *RestServer) handle(writer http.ResponseWriter, req *http.Request) {
	var path = req.URL.Path

	// 是否显示Pretty后的JSON
	var shouldPretty = req.Header.Get("X-Edge-Response-Pretty") == "on"

	// 兼容老的Header
	var oldShouldPretty = req.Header.Get("Edge-Response-Pretty")
	if len(oldShouldPretty) > 0 {
		shouldPretty = oldShouldPretty == "on"
	}

	// 欢迎页
	if path == "/" {
		this.writeJSON(writer, maps.Map{
			"code":    200,
			"message": "Welcome to API",
			"data":    maps.Map{},
		}, shouldPretty)
		return
	}

	var matches = servicePathReg.FindStringSubmatch(path)
	if len(matches) != 3 {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	var serviceName = matches[1]
	var methodName = matches[2]

	serviceType, ok := restServicesMap[serviceName]
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if len(methodName) == 0 {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	// 再次查找
	methodName = strings.ToUpper(string(methodName[0])) + methodName[1:]
	var method = serviceType.MethodByName(methodName)
	if !method.IsValid() {
		// 兼容Enabled
		if strings.Contains(methodName, "Enabled") {
			methodName = strings.Replace(methodName, "Enabled", "", 1)
			method = serviceType.MethodByName(methodName)
			if !method.IsValid() {
				writer.WriteHeader(http.StatusNotFound)
				return
			}
		} else {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
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
	var ctx = context.Background()

	if serviceName != "APIAccessTokenService" || (methodName != "GetAPIAccessToken" && methodName != "getAPIAccessToken") {
		// 校验TOKEN
		var token = req.Header.Get("X-Edge-Access-Token")
		if len(token) == 0 {
			token = req.Header.Get("Edge-Access-Token")
			if len(token) == 0 {
				this.writeJSON(writer, maps.Map{
					"code":    400,
					"data":    maps.Map{},
					"message": "require 'X-Edge-Access-Token' header",
				}, shouldPretty)
				return
			}
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
		} else if accessToken.AdminId > 0 {
			ctx = rpcutils.NewPlainContext("admin", int64(accessToken.AdminId))
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

	// TODO 可以设置最大可接收内容尺寸
	body, err := io.ReadAll(io.LimitReader(req.Body, 32*sizes.M))
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(err.Error()))
		return
	}

	// 如果为空，表示传的数据为空
	if len(body) == 0 {
		body = []byte("{}")
	}

	// 请求数据
	var reqValue = reflect.New(method.Type().In(1).Elem()).Interface()
	err = json.Unmarshal(body, reqValue)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte("Decode request failed: " + err.Error() + ". Request body should be a valid JSON data"))
		return
	}

	var result = method.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(reqValue)})
	var resultErr = result[1].Interface()
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
		var data = maps.Map{
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
