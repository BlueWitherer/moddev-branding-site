package access

import (
	"fmt"
	"net/http"
)

func GetDomain(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

func FullURL(r *http.Request) string {
	base := GetDomain(r)
	return fmt.Sprintf("%s%s", base, r.RequestURI)
}
