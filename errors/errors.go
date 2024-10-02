package errors

import (
	errs "errors"
)

var (
	ErrorBadResponse = errs.New("Неверный ответ от https://hmtpk.ru")
	ErrorBadRequest  = errs.New("Неверный запрос")
)
