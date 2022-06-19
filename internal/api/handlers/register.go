package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sergeysynergy/hardtest/internal/gophermart"
)

func (h *handler) register(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct != ContentTypeApplicationJSON {
		err := fmt.Errorf("wrong content type, JSON needed")
		h.error(w, r, err, http.StatusBadRequest)
		return
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to read request body - %w", err), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var creds gophermart.Credentials
	err = json.Unmarshal(reqBody, &creds)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to unmarshal body - %w", err), http.StatusBadRequest)
		return
	}

	session, err := h.gm.Register(&creds)
	if err != nil {
		msg := "failed to register new user"
		if errors.Is(err, gophermart.ErrLoginAlreadyTaken) {
			h.error(w, r, fmt.Errorf("%s - %w", msg, err), http.StatusConflict)
			return
		}
		h.error(w, r, fmt.Errorf("%s - %w", msg, err), http.StatusInternalServerError)
		return
	}
	if session == nil {
		h.error(w, r, fmt.Errorf("got nil session"), http.StatusInternalServerError)
		return
	}

	// создадим куку со сроком годности
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   session.Token,
		Expires: session.Expiry,
	})

	msg := fmt.Sprintf("session for user `%s` successfully created", creds.Login)
	h.log(r, LogLvlDebug, msg)
}
