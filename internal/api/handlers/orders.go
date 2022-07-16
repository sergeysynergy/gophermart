package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sergeysynergy/gophermart/internal/gophermart"
	"io/ioutil"
	"net/http"
	"strconv"
)

func (h *handler) postOrders(w http.ResponseWriter, r *http.Request) {
	var err error

	ct := r.Header.Get("Content-Type")
	if ct != ContentTypeTextPlain {
		err = fmt.Errorf("wrong content type, %s needed", ContentTypeTextPlain)
		h.error(w, r, err, http.StatusBadRequest)
		return
	}

	c, err := h.authCheck(w, r)
	if err != nil {
		// 401 — пользователь не авторизован
		return
	}

	u, err := h.gm.Users.Get(c.UserID)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to get user by ID - %w", err), http.StatusInternalServerError)
		return
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to read request body - %w", err), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	orderID, err := strconv.Atoi(string(reqBody))
	if err != nil {
		h.error(w, r, fmt.Errorf("%s - %w", gophermart.ErrOrderInvalidFormat, err), http.StatusUnprocessableEntity)
		return
	}

	err = h.gm.PostOrders(uint64(orderID), u.ID)
	if err != nil {
		// 200 — номер заказа уже был загружен этим пользователем
		if errors.Is(err, gophermart.ErrOrderAlreadyLoadedByUser) {
			w.WriteHeader(http.StatusOK)
			msg := fmt.Sprintf("order %d has already been uploaded by this user", orderID)
			h.log(r, LogLvlInfo, msg)
			return
		}

		// 409 — номер заказа уже был загружен другим пользователем
		if errors.Is(err, gophermart.ErrOrderAlreadyLoadedByAnotherUser) {
			h.error(w, r, gophermart.ErrOrderAlreadyLoadedByAnotherUser, http.StatusConflict)
			return
		}

		// 422 — неверный формат номера заказа
		if errors.Is(err, gophermart.ErrOrderInvalidFormat) {
			h.error(w, r, gophermart.ErrOrderInvalidFormat, http.StatusUnprocessableEntity)
			return
		}

		// 500 — внутренняя ошибка сервера
		h.error(w, r, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	msg := fmt.Sprintf("new order %d has been accepted for processing", orderID)
	h.log(r, LogLvlInfo, msg)
}

func (h *handler) getOrders(w http.ResponseWriter, r *http.Request) {
	var err error

	c, err := h.authCheck(w, r)
	if err != nil {
		// 401 — пользователь не авторизован
		return
	}
	u, err := h.gm.Users.Get(c.UserID)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to get user by ID - %w", err), http.StatusInternalServerError)
		return
	}

	userID := u.ID

	// БЛИНСКИЙ ЁЖ, ТЗ НАС КРУПНО НАДУЛО: там указано `Content-Length: 0`
	//cl := r.Header.Get("Content-Length")
	//if cl != "0" {
	//	err = fmt.Errorf("wrong content length")
	//	h.error(w, r, err, http.StatusBadRequest)
	//	return
	//}

	proxyOrders, err := h.gm.GetOrders(userID)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to get all orders - %w", err), http.StatusInternalServerError)
		return
	}

	if len(proxyOrders) == 0 {
		h.error(w, r, fmt.Errorf("orders not found for this user"), http.StatusNoContent)
		return
	}

	body, err := json.Marshal(&proxyOrders)
	if err != nil {
		h.error(w, r, fmt.Errorf("failed to marshal JSON - %w", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.Write(body)
}
