package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func (h *handler) error(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	reqID := middleware.GetReqID(r.Context())

	type errorJSON struct {
		Error      string
		StatusCode int
	}
	e := errorJSON{
		Error:      err.Error(),
		StatusCode: statusCode,
	}

	prefix := "[ERROR]"
	if reqID != "" {
		prefix = fmt.Sprintf("[%s] [ERROR]", reqID)
	}

	b, errMarshal := json.Marshal(e)
	if errMarshal != nil {
		msg := fmt.Sprintf(`{"Error": "Failed to marshal error - %s", "StatusCode": 500`, err)
		w.Write([]byte(msg))
		log.Println(prefix, msg)
		return
	}

	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(statusCode)
	w.Write(b)
	log.Println(prefix, e)
}
