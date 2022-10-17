package utl

import (
	"net"
	"net/http"
)

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
	XUserId       = "X-User-Id"
)

func GetRemoteIp(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get(XRealIP); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get(XForwardedFor); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}

func GetUserId(req *http.Request) string {
	uid := req.Header.Get(XUserId)
	if uid == "" {
		uid = req.FormValue("uid")
	}
	if uid == "" {
		return GetRemoteIp(req)
	} else {
		return uid
	}
}

