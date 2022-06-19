package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (h *handler) getBalance(w http.ResponseWriter, r *http.Request) {
	var err error

	// и тут дурят нашего брата? видимо, я не так понимаю запись в ТЗ `Content-Length: 0`
	//cl := r.Header.Get("Content-Length")
	//if cl != "0" {
	//	err = fmt.Errorf("wrong content length")
	//	h.error(w, r, err, http.StatusBadRequest)
	//	return
	//}

	c, err := h.authCheck(w, r)
	if err != nil {
		// 401 — пользователь не авторизован
		return
	}
	user, err := h.gm.Users.Get(c.UserID)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to get user by ID - %w", err), http.StatusInternalServerError)
		return
	}

	balanceProxy, err := h.gm.GetBalance(user.ID)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to get balance for user `%s` - %w", user.Login, err), http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(&balanceProxy)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to marshal JSON - %w", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.Write(body)
}
