package router

import (
	"net"
	"net/http"

	"github.com/nekr0z/muhame/internal/httpclient"
)

func trusted(subnet string) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.Header.Get(httpclient.HeaderRealIP)
			if !isInSubnet(ip, subnet) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isInSubnet(ipStr, cidrStr string) bool {
	_, cidrNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	return cidrNet.Contains(ip)
}
