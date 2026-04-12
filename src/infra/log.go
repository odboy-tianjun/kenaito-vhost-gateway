package infra

import (
	"log"
	"net/http"
	"time"
)

func LogRequest(r *http.Request, prefix string, status int, size int64, start time.Time) {
	duration := time.Since(start)
	log.Printf("%s - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" prefix=%s %.3f ms",
		r.RemoteAddr,
		time.Now().Format("02/Jan/2006:15:04:05 -0700"),
		r.Method,
		r.URL.Path,
		r.Proto,
		status,
		size,
		r.Header.Get("Referer"),
		r.Header.Get("User-Agent"),
		prefix,
		float64(duration.Microseconds())/1000.0,
	)
}
