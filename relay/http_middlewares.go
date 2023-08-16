package relay

import (
	"compress/gzip"
	"net/http"
	"net/http/httputil"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"time"
)

func allMiddlewares(h *HTTP, handlerFunc relayHandlerFunc) relayHandlerFunc {
	var res = handlerFunc
	for _, middleware := range middlewares {
		res = middleware(h, res)
	}

	return res
}

func (h *HTTP) logMiddleWare(next relayHandlerFunc) relayHandlerFunc {
	return relayHandlerFunc(func(h *HTTP, w http.ResponseWriter, r *http.Request, start time.Time) {
		if h.log {
			h.logger.Println("got request from: " + r.RemoteAddr + " on: " + r.URL.Path)

			// DEBUG
			h.logger.Printf("\nDEBUG\nr.URL.Host   : %s \nr.URL.Path   : %s \nr.URL.RawPath: %s \nr.Host       : %s \nr.RemoteAddr : %s \nr.RequestURI : %s \n", r.URL.Host, r.URL.Path, r.URL.RawPath, r.Host, r.RemoteAddr, r.RequestURI)
			reqDump, err := httputil.DumpRequest(r, true)
			if err != nil {
				h.logger.Fatal(err)
			}
			h.logger.Printf("REQUEST:\n%s", string(reqDump))
		}

		var bodyBytes []byte
		var err error

		if r.Body != nil {
			bodyBytes, err = ioutil.ReadAll(r.Body)
			if err != nil {
				h.logger.Printf("Body reading error: %v", err)
			}
			defer r.Body.Close()
		}

		h.logger.Printf("Headers: %+v\n", r.Header)

		if len(bodyBytes) > 0 {
			var prettyJSON bytes.Buffer
			if err = json.Indent(&prettyJSON, bodyBytes, "", "\t"); err != nil {
				h.logger.Printf("JSON parse error: %v", err)
			}
			h.logger.Println(string(prettyJSON.Bytes()))
		} else {
			h.logger.Printf("Body: No Body Supplied\n")
		}

		next(h, w, r, start)
	})
}

func (h *HTTP) bodyMiddleWare(next relayHandlerFunc) relayHandlerFunc {
	return relayHandlerFunc(func(h *HTTP, w http.ResponseWriter, r *http.Request, start time.Time) {
		var body = r.Body

		if r.Header.Get("Content-Encoding") == "gzip" {
			b, err := gzip.NewReader(r.Body)
			if err != nil {
				jsonResponse(w, response{http.StatusBadRequest, "unable to decode gzip body"})
				return
			}
			defer b.Close()
			body = b
		}

		r.Body = body
		next(h, w, r, start)
	})
}

func (h *HTTP) queryMiddleWare(next relayHandlerFunc) relayHandlerFunc {
	return relayHandlerFunc(func(h *HTTP, w http.ResponseWriter, r *http.Request, start time.Time) {
		queryParams := r.URL.Query()

		if queryParams.Get("db") == "" && (r.URL.Path == "/write" || r.URL.Path == "/api/v1/prom/write") {
			jsonResponse(w, response{http.StatusBadRequest, "missing parameter: db"})
			return
		}

		if queryParams.Get("rp") == "" && h.rp != "" {
			queryParams.Set("rp", h.rp)
		}

		r.URL.RawQuery = queryParams.Encode()
		next(h, w, r, start)
	})

}

func (h *HTTP) rateMiddleware(next relayHandlerFunc) relayHandlerFunc {
	return relayHandlerFunc(func(h *HTTP, w http.ResponseWriter, r *http.Request, start time.Time) {
		if h.rateLimiter != nil && !h.rateLimiter.Allow() {
			jsonResponse(w, response{http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests)})
			return
		}

		next(h, w, r, start)
	})
}
