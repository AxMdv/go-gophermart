package app_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/AxMdv/go-gophermart/internal/app"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E(t *testing.T) {

	// init app
	go func() {
		app, err := app.New()
		require.NoError(t, err, "Error during launching app")

		err = app.Run()
		require.NoError(t, err, "Error during running app")
	}()
	time.Sleep(5 * time.Second)
	httpc := resty.New().
		SetBaseURL("http://localhost:8080")

	m := []byte(`
	{
		"login": "asd1233",
		"password": "zxc"
	} 
	`)
	req := httpc.R().
		SetHeader("Content-Type", "application/json").
		SetBody(m)

	resp, err := req.Post("/api/user/register")

	assert.NoError(t, err, "Ошибка при попытке сделать запрос на регистрацию пользователя в системе лояльности")
	assert.Equal(t, http.StatusOK, resp.StatusCode(), "Несоответствие статус кода ответа ожидаемому в хендлере")

	setCookieHeader := resp.Header().Get("Set-Cookie")
	assert.True(t, setCookieHeader != "", "Не удалось обнаружить авторизационные данные в ответе")

}
