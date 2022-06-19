package handlers

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sergeysynergy/hardtest/internal/db"
	"github.com/sergeysynergy/hardtest/internal/gophermart"
)

func TestRegister(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name     string
		addr     string
		user     gophermart.Credentials
		notClear bool
		want     want
	}{
		{
			name: "status Ok",
			addr: "/api/user/register",
			user: gophermart.Credentials{
				Login:    "testov",
				Password: "Passw0rd33",
			},
			notClear: true,
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "status conflict: login already taken",
			addr: "/api/user/register",
			user: gophermart.Credentials{
				Login:    "testov",
				Password: "Passw0rd33",
			},
			notClear: true,
			want: want{
				statusCode: http.StatusConflict,
			},
		},
		{
			name: "status unauthorized",
			addr: "/api/user/login",
			user: gophermart.Credentials{
				Login:    "unknownUser",
				Password: "Passw0rd33",
			},
			notClear: true,
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "status unauthorized",
			addr: "/api/user/login",
			user: gophermart.Credentials{
				Login:    "testov",
				Password: "wrongPass",
			},
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st, err := db.New("user=postgres password=Passw0rd33 host=localhost port=5432 dbname=gophermarttest")
			require.NoError(t, err)

			gm := gophermart.New(st)
			h := New(gm)
			ts := httptest.NewServer(h.GetRouter())
			defer ts.Close()

			client := resty.New()
			resp, err := client.R().
				EnableTrace().
				SetHeader("Content-Type", ContentTypeApplicationJSON).
				SetBody(tt.user).
				Post(ts.URL + tt.addr)

			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode())

			fmt.Println("::: body:", string(resp.Body()))

			if !tt.notClear {
				err = gm.Users.Delete(tt.user.Login)
				require.NoError(t, err)
			}
		})
	}
}
