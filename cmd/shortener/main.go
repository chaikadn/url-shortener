package main

import (
	"log"

	"github.com/chaikadn/url-shortener/internal/app/config"
	"github.com/chaikadn/url-shortener/internal/app/handler"
	"github.com/chaikadn/url-shortener/internal/app/logger"
	"github.com/chaikadn/url-shortener/internal/app/server"
	"github.com/chaikadn/url-shortener/internal/app/storage"
	"go.uber.org/zap"
)

/*
Эндпоинты:
--------------------------------------------------------------------
POST / - сокращает длинный URL из тела запроса в формате text/plain
	201 - создан новый короткий URL
	Тело ответа - новый короткий URL в формате text

	406 - длинный URL из тела запроса уже есть
	Тело ответа - существующий короткий URL, соответствующий длинному URL из тела запроса

	500 - если созданный новый короткий URL оказался равен уже существующему
	Тело ответа - failed to get short url

	400 - если не удалось прочитать тело запроса
	Тело ответа - failed to read body

	400 - если не получилось записать короткий URL в тело ответа
	Тело ответа - failed to write body

	400 - остальные ошибки (URL не валидный и т.д.)
	Тело ответа - failed to shorten url
--------------------------------------------------------------------
GET /ping - пинг хранилища
	200 - Хранилище работает корректно
	Тела ответа нет

	500 - пинг не удался
	Тело ответа - failed to connect to storage
--------------------------------------------------------------------
GET /{short-url} - перенаправляет на сайт с длинным URL, соответствующим short-url
	307 - перенаправление на длинный URL, соответствующий короткому {short-url}
	Заголовок Location: длинный URl
	Тела ответа нет

	404 - длинный URL не найден по такому короткому
	Тело ответа - url not found

	500 - остальные ошибки
	Тело ответа - failed to get url
--------------------------------------------------------------------
POST /api/shorten - сокращает длинный URL из тела запроса в формате JSON {"url": "длинный URL"}
	201 - создан новый короткий URL
	Тело ответа - новый короткий URL в JSON формате {"result": "короткий URL"}

	406 - длинный URL из тела запроса уже есть
	Тело ответа - существующий короткий URL, соответствующий длинному URL в JSON формате {"result": "короткий URL"}

	500 - если созданный новый короткий URL оказался равен уже существующему
	Тело ответа - failed to get short url

	400 - если не удалось прочитать тело запроса
	Тело ответа - failed to decode json body

	400 - если не получилось закодировать короткий URL в JSON в тело ответа
	Тело ответа - failed to encode json

	400 - остальные ошибки (URL не валидный и т.д.)
	Тело ответа - failed to shorten url
--------------------------------------------------------------------
POST /api/shorten/batch - сокращает пачку длинных URL из тела запроса в формате JSON:
	[
		{
			"correlation_id": "айди пары 1",
			"original_url": "длинный URL 1"
		},
		{
			"correlation_id": "айди пары 2",
			"original_url": "длинный URL 2"
		},
		...
	]

	201 - создана пачка коротких URL
	Тело ответа - пачка сокращенных URL:
	[
		{
			"correlation_id": "айди пары 1",
			"short_url": "короткий URL 1"
		},
		{
			"correlation_id": "айди пары 2",
			"short_url": "короткий URL 2"
		},
		...
	]

	406 - ???

	400 - если не удалось декодировать JSON
	Тело ответа - failed to decode json body

	400 - если пачка пустая
	Тело ответа - batch cannot be empty

	400 - если не получилось закодировать короткие URL в JSON в тело ответа
	Тело ответа - failed to encode json

	400 - остальные ошибки
	Тело ответа - failed to shorten batch
--------------------------------------------------------------------
GET /api/user/urls - выводит список URL, сокращенных пользователем, тела запроса нет
Заголовок "???": "???"

	200 - у пользователя с данным ID нашлись сохраненные URL
	Тело ответа - список пар длиннный-короткий URL в формате JSON:
	[
		{
			"short_url": "короткий URL 1",
			"original_url": "длинный URL 1"
		},
			"short_url": "короткий URL 2",
			"original_url": "длинный URL 2"
		},
		...
	]

	204 - Пользователь с данным ID еще ничего не сохранял
	Тела ответа нет

	401 - ошибка авторизации пользователя, ID пользователя пустой или невалидный
	Тело ответа - authorization failed

	Прочие ошибки
--------------------------------------------------------------------
*/

func main() {
	cfg := config.New()
	if err := cfg.Load(); err != nil {
		log.Fatalf("\t\tfailed to initialize config: %v", err)
	}

	zLog, err := logger.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("\t\tfailed to initialize config: %v", err)
	}
	defer zLog.Sync()

	stg, err := storage.New(cfg, zLog)
	if err != nil {
		zLog.Fatal("failed to initialize storage", zap.Error(err))
	}
	defer stg.Close()

	hnd, err := handler.New(cfg, zLog, stg)
	if err != nil {
		zLog.Fatal("failed to initialize handler", zap.Error(err))
	}

	srv := server.New(cfg, zLog, hnd)

	zLog.Info("Starting server", zap.String("host", cfg.Host))
	if err := srv.ListenAndServe(); err != nil {
		zLog.Fatal("failed to start server", zap.Error(err))
	}
}
