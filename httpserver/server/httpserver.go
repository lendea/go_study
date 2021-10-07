package server

import (
	"context"
	"github.com/felixge/httpsnoop"
	"net/http"
	"strings"
	"time"
	"turtorial.lendea.cn/common/logger"
)

func MakeHTTPServer(ctx context.Context, version string) *http.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
				<head><title> httpserver study</title></head>
				<body>
				<h1>HttpServer Study Health URL</h1>
				<p><a href="` + "healthz" + `">HealthCheck</a></p>
				</body>
				</html>`))
		if err != nil {
			logger.For(ctx).Errorf("failed handling writer,error:%v", err)
		}
	})

	// health endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
	})

	var handler http.Handler = mux
	// 日志记录器包装 mux
	handler = requestHandler(ctx, handler, version)
	srv := &http.Server{
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      handler,
	}
	return srv
}

func requestHandler(ctx context.Context, h http.Handler, version string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		//rewrite header,将request 中带的 header 写入 response header
		w.Header().Add("version", version)
		for key, reqHeader := range r.Header {
			w.Header().Add(key, strings.Join(reqHeader, ","))
		}

		//封装请求和响应信息。
		ri := &HTTPReqInfo{
			method:    r.Method,
			uri:       r.URL.String(),
			referer:   r.Header.Get("Referer"),
			userAgent: r.Header.Get("User-Agent"),
		}

		ri.ipaddr = requestGetRemoteAddress(r)
		// 这里运行处理器 h 并捕获有关 HTTP 请求的信息
		m := httpsnoop.CaptureMetrics(h, w, r)
		ri.code = m.Code
		ri.size = m.Written
		ri.duration = m.Duration
		logger.For(ctx).Infof("Http Request Info, [Method=%s], [URL=%s], [Referer=%s], [UserAgent=%s], [IpAddr=%s], [Code=%d], [WrittenByteSize=%d], [Duration=%d]",
			ri.method, ri.uri, ri.referer, ri.userAgent, ri.ipaddr, ri.code, ri.size, ri.duration)
	}
	// 用 http.HandlerFunc 包装函数，这样就实现了 http.Handler 接口
	return http.HandlerFunc(fn)
}

// Request.RemoteAddress 包含了端口，我们需要把它删掉，比如: "[::1]:58292" => "[::1]"
func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

// requestGetRemoteAddress 返回发起请求的客户端 ip 地址，这是出于存在 http 代理的考量
func requestGetRemoteAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIP := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIP == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For 可能是以","分割的地址列表
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		return parts[0]
	}
	return hdrRealIP
}

//HTTPReqInfo 描述了http相关信息
type HTTPReqInfo struct {
	// GET 等方法
	method  string
	uri     string
	referer string
	ipaddr  string
	// 响应状态码，如 200，204
	code int
	// 所发送响应的字节数
	size int64
	// 处理花了多长时间
	duration  time.Duration
	userAgent string
}
