package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"
)

//// HTTP_VERBS
func (a *app) Get(url string, handler func(http.ResponseWriter, *http.Request)) {
	a.routes["GET"] = append(a.routes["GET"], url)
	a.buildParams("GET", url)
	a.HandlerRegister(url, "GET", handler)
}
func (a *app) Post(url string, handler func(http.ResponseWriter, *http.Request)) {
	a.routes["POST"] = append(a.routes["POST"], url)
	a.buildParams("POST", url)
	a.HandlerRegister(url, "POST", handler)
}
func (a *app) Put(url string, handler func(http.ResponseWriter, *http.Request)) {
	a.routes["PUT"] = append(a.routes["PUT"], url)
	a.buildParams("PUT", url)
	a.HandlerRegister(url, "PUT", handler)
}
func (a *app) Patch(url string, handler func(http.ResponseWriter, *http.Request)) {
	a.routes["PATCH"] = append(a.routes["PATCH"], url)
	a.buildParams("PATCH", url)
	a.HandlerRegister(url, "PATCH", handler)
}
func (a *app) Delete(url string, handler func(http.ResponseWriter, *http.Request)) {
	a.routes["DELETE"] = append(a.routes["DELETE"], url)
	a.buildParams("DELETE", url)
	a.HandlerRegister(url, "DELETE", handler)
}

//// MAIN  MIDDLEWARE
func (a *app) Middleware(methods []string, route string, handler func(http.ResponseWriter, *http.Request) (*http.Request, bool, []byte, int)) {
	for _, method := range methods {
		if _, ok := a.middleware[route]; !ok {
			a.middleware[route] = make(map[string][]func(http.ResponseWriter, *http.Request) (*http.Request, bool, []byte, int))
		}
		if _, ok := a.middleware[route][method]; !ok {
			a.middleware[route][method] = make([]func(http.ResponseWriter, *http.Request) (*http.Request, bool, []byte, int), 0)
		}
		a.middleware[route][method] = append(a.middleware[route][method], handler)
	}
}
func (a *app) methodMiddleware(url string, method string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := &r
		if _, ok := a.middleware[url]; ok {
			if _, ok := a.middleware[url][method]; ok {
				for _, mw := range a.middleware[url][method] {
					r, passNext, data, status := mw(w, *s)
					s = &r
					if !passNext {
						Response(r, w, data, status)
						return
					}
				}
			}
		}
		handler.ServeHTTP(w, *s)
	})
}
func (a *app) UrlNotFound(r *http.Request, w http.ResponseWriter, method string, url string, originalURL string) bool {
	if ok := a.contains(method, url); !ok {
		_url := fmt.Sprintf(`{"error":"URL %s was not found on the server"}`, url)
		message := []byte(_url)
		Response(r, w, message, 404)
		return false
	}

	return true
}
func (a *app) UrlParser(url string, originalURL string, http_method string) (map[string]string, bool) {
	_url := strings.Split(url, "/")
	index := 0
	_index := 0
	requestParams := make(map[string]string)
	var track_params []int
	if len(_url) == len(a.params[http_method][originalURL]) {
		for i, key := range a.params[http_method][originalURL] {
			if !strings.Contains(_url[i], "{") {
				if _url[i] == key {
					index++
				}
			} else {
				_index++
				track_params = append(track_params, i)
			}
		}
		fmt.Println(track_params)
		if _index > 0 {
			for i, _ := range track_params {

				tempParam := strings.Split(a.params[http_method][url][track_params[i]], "{")[1]
				tempParam = strings.Split(tempParam, "}")[0]
				requestParams[tempParam] = _url[track_params[i]]
			}
		}
	}

	return requestParams, len(requestParams) > 0
}
func (a *app) HandlerRegister(url string, method string, handler func(http.ResponseWriter, *http.Request)) {
	if _, ok := a.handlers[url]; !ok {
		a.handlers[url] = make(map[string]func(http.ResponseWriter, *http.Request))
	}
	if _, ok := a.handlers[url][method]; !ok {
		a.handlers[url][method] = handler
	}
}

//// SERVER MAIN HANDLERS
func (a *app) ServeHttpPrimary(w http.ResponseWriter, r *http.Request, url string, method string) {
	if _, ok := a.handlers[url]; ok {
		if _, ok := a.handlers[url][method]; ok {
			executable := a.handlers[url][method]
			a.methodMiddleware(url, method, http.HandlerFunc(executable)).ServeHTTP(w, r)
		} else {
			MethodGuard(w, r)
			return
		}
	} else {
		a.UrlNotFound(r, w, method, url, url)
		return
	}
}
func (a *app) ServeHttpSecondary(w http.ResponseWriter, r *http.Request, _url string, method string) {
	__url, key_params := a.RegexpRouting(_url, method)
	url := strings.Split(__url, "?")[0]
	if _, ok := a.handlers[url]; ok && len(key_params) > 0 {
		if _, ok := a.handlers[url][method]; ok {
			executable := a.handlers[url][method]
			// fmt.Println("name:", GetFunctionName(executable))
			params := RequestQuery(r)
			for key := range key_params {
				params.Set(key, key_params[key])
			}
			r.URL.RawQuery = params.Encode()
			fmt.Println(url)
			a.methodMiddleware(url, method, http.HandlerFunc(executable)).ServeHTTP(w, r)
		} else {
			MethodGuard(w, r)
			return
		}
	} else {
		a.UrlNotFound(r, w, method, url, url)
		return
	}
}

/// HTTP MODULE HELPERS

func MethodGuard(w http.ResponseWriter, r *http.Request) bool {
	message := []byte(`{"error":"Method not allowed"}`)
	Response(r, w, message, 405)
	return false
}

func HttpConsoleLogger(url string, method string, ip string, status int) {
	http_verbs := map[int]string{
		400: "Bad Request",
		401: "Unauthorized",
		403: "Forbidden",
		404: "Not found",
		405: "Method Not Allowed",
		415: "Unsupported Media Type",
		500: "Internal Server Error",
		503: "Service Unavailable",
		504: "Gateway Timeout",
		200: "OK",
		201: "Created",
		202: "Accepted",
		204: "No Content",
	}
	now := time.Now()
	fmt.Printf(" [%s]: (%s) from remote address (ip): %s at %s , %d %s\n", method, url, ip, now, status, http_verbs[status])
}

func (a *app) HttpFileLogger(w http.ResponseWriter, r *http.Request) {

}

/// HTTP UTILS
func (a *app) RegexpRouting(url string, http_method string) (string, map[string]string) {
	main_expression := `[a-zA-Z0-9]+/?`
	currentUrlContent := strings.Split(url, "/")
	for key, _ := range a.secondaryRoutes[http_method] {
		urls := strings.Split(key, "/")
		key_params := make(map[string]string)
		if len(urls) == len(currentUrlContent) {
			for i, _url := range urls {
				if strings.Contains(_url, "{") && strings.Contains(_url, "}") {
					key_params[_url[:len(_url)-1][1:]] = currentUrlContent[i]
					urls[i] = main_expression
				} else {
					if currentUrlContent[i] != _url {
						return "", make(map[string]string)
					}
				}
			}
			build := strings.Join(urls, "/")
			reg := regexp.MustCompile(build)
			check := reg.MatchString(url)
			if check {
				// fmt.Printf("found key %s for case: %s \n", key, url)
				build_query_params := "?params=true"
				for p := range key_params {
					build_query_params += fmt.Sprintf("&%s=%s", p, key_params[p])
				}
				// fmt.Println(key + build_query_params)
				return key, key_params
			}
		}
	}
	// reg := regexp.MustCompile(`{[a-zA-Z]*}`)
	// check := reg.MatchString(url)
	// fmt.Println(check)
	return "", make(map[string]string)
}

func (a *app) buildParams(http_method string, url string) {
	if _, ok := a.params[http_method]; !ok {
		a.params[http_method] = make(map[string][]string)
	}
	if _, ok := a.primeRoutes[http_method]; !ok {
		a.primeRoutes[http_method] = make(map[string]bool)
	}
	if _, ok := a.primeRoutes[http_method][url]; !ok && !strings.Contains(url, "{") {
		a.primeRoutes[http_method][url] = true
	} else {
		if _, ok := a.secondaryRoutes[http_method]; !ok {
			a.secondaryRoutes[http_method] = make(map[string]bool)
		}
		a.secondaryRoutes[http_method][url] = true
		return
	}
	a.params[http_method][url] = strings.Split(url, "/")
	return
}

func RequestQuery(r *http.Request) url.Values {
	return r.URL.Query()
}

func GetHeader(r *http.Request, key string) string {
	header := r.Header.Get(key)
	return header
}

func RequestBody(r *http.Request, receiver interface{}) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(reqBody, &receiver)
}

func setHeader(w http.ResponseWriter, key string, value string) {
	w.Header().Set(key, value)
}

func ResponseLogger(r *http.Request, status int) {
	ip := requestGetRemoteAddress(r)
	method := r.Method
	url := r.URL.Path
	HttpConsoleLogger(url, method, ip, status)
}

func Response(r *http.Request, w http.ResponseWriter, value []byte, status int) {
	w.WriteHeader(status)
	w.Write(value)
	ResponseLogger(r, status)
}

func JsonResponse(r *http.Request, w http.ResponseWriter, data interface{}, status int) {
	json.NewEncoder(w).Encode(data)
	ResponseLogger(r, status)
}

func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

func requestGetRemoteAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIP := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIP == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with ","
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		// TODO: should return first non-local address
		return parts[0]
	}
	return hdrRealIP
}

/// CONTEXT HANDLERS
func GetContextValue(r *http.Request, value interface{}) interface{} {
	return r.Context().Value(value)
}
func SetContextValue(r *http.Request, key interface{}, value interface{}) *http.Request {
	ctx := context.WithValue(r.Context(), key, value)
	return r.WithContext(ctx)
}

/// HELPERS

func (a *app) contains(method string, url string) bool {
	for _, key := range a.routes[method] {
		if key == url {
			return true
		}
	}
	return false
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
