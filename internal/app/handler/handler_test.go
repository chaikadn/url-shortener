package handler

// func TestHandler_postURL(t *testing.T) {
// 	storage := memory.NewStorage()
// 	handler := New(storage, &config.Config{})

// 	r := chi.NewRouter()
// 	r.Post("/", handler.postURL)
// 	server := httptest.NewServer(r)
// 	defer server.Close()

// 	// TODO: mock random url generator
// 	tests := []struct {
// 		name         string
// 		method       string
// 		body         string
// 		expectedCode int
// 		// expectedBody     string
// 	}{
// 		{
// 			name:         "positive test",
// 			method:       http.MethodPost,
// 			body:         "https://practicum.yandex.ru/",
// 			expectedCode: http.StatusCreated,
// 			// expectedBody: "fixedBody",
// 		},
// 		{
// 			name:         "empty body test",
// 			method:       http.MethodPost,
// 			body:         "",
// 			expectedCode: http.StatusBadRequest,
// 		},
// 		{
// 			name:         "invalid url test",
// 			method:       http.MethodPost,
// 			body:         "invalid-url",
// 			expectedCode: http.StatusBadRequest,
// 		},
// 		{
// 			name:         "wrong method test",
// 			method:       http.MethodGet,
// 			body:         "any-url",
// 			expectedCode: http.StatusMethodNotAllowed,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			req := resty.New().R()
// 			req.Method = tt.method
// 			req.URL = server.URL
// 			req.Body = tt.body

// 			resp, err := req.Send()
// 			assert.NoError(t, err, "error making HTTP request")

// 			assert.Equal(t, tt.expectedCode, resp.StatusCode(), "Response code didn't match expected")
// 			// assert.Equal(t, tt.expectedBody, string(resp.Body()))
// 		})
// 	}
// }

// func TestHandler_getURL(t *testing.T) {
// 	storage := memory.NewStorage()
// 	handler := New(storage, &config.Config{})

// 	r := chi.NewRouter()
// 	r.Get("/{id}", handler.getURL)
// 	server := httptest.NewServer(r)
// 	defer server.Close()

// 	id, _ := storage.Add("https://practicum.yandex.ru/")

// 	tests := []struct {
// 		name             string
// 		method           string
// 		path             string
// 		expectedCode     int
// 		expectedBody     string
// 		expectedLocation string
// 	}{
// 		{
// 			name:             "positive test",
// 			method:           http.MethodGet,
// 			path:             "/" + id,
// 			expectedCode:     http.StatusTemporaryRedirect,
// 			expectedLocation: "https://practicum.yandex.ru/",
// 		},
// 		{
// 			name:         "non-existent URL",
// 			method:       http.MethodGet,
// 			path:         "/invalid-id",
// 			expectedCode: http.StatusBadRequest,
// 			expectedBody: "Unable to get URL\n",
// 		},
// 		{
// 			name:         "wrong method",
// 			method:       http.MethodPost,
// 			path:         "/any-id",
// 			expectedCode: http.StatusMethodNotAllowed,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			req := resty.New().SetRedirectPolicy(resty.NoRedirectPolicy()).R() // disable autoredirect
// 			req.Method = tt.method
// 			req.URL = server.URL + tt.path

// 			resp, err := req.Send()

// 			// ignore redirect error
// 			if errors.Unwrap(err) == resty.ErrAutoRedirectDisabled {
// 				err = nil
// 			}
// 			assert.NoError(t, err, "error making HTTP request")

// 			assert.Equal(t, tt.expectedCode, resp.StatusCode(), "Response code didn't match expected")
// 			assert.Equal(t, tt.expectedLocation, resp.Header().Get("Location"))
// 			assert.Equal(t, tt.expectedBody, string(resp.Body()))
// 		})
// 	}
// }

// TODO
// func TestHandler_shorten(t *testing.T) {
// 	storage := memory.NewStorage()
// 	handler := New(storage, &config.Config{})

// 	// REFACTOR
// 	r := chi.NewRouter()
// 	r.Post("/", handler.shorten)
// 	server := httptest.NewServer(r)
// 	defer server.Close()

// 	tests := []struct {
// 		name         string
// 		method       string
// 		body         string
// 		expectedCode int
// 	}{
// 		{
// 			name:         "positive test",
// 			method:       http.MethodPost,
// 			body:         `{"url": "https://practicum.yandex.ru"}`,
// 			expectedCode: http.StatusCreated,
// 		},
// 		{
// 			name:         "empty body test",
// 			method:       http.MethodPost,
// 			expectedCode: http.StatusInternalServerError,
// 		},
// 		{
// 			name:         "invalid url test",
// 			method:       http.MethodPost,
// 			body:         `{"url": "invalid-url"}`,
// 			expectedCode: http.StatusBadRequest,
// 		},
// 		{
// 			name:         "wrong method test",
// 			method:       http.MethodGet,
// 			body:         `{}`,
// 			expectedCode: http.StatusMethodNotAllowed,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			req := resty.New().R()
// 			req.Method = tt.method
// 			req.URL = server.URL
// 			if len(tt.body) > 0 {
// 				req.SetHeader("Content-Type", "application/json")
// 				req.SetBody(tt.body)
// 			}

// 			resp, err := req.Send()
// 			assert.NoError(t, err, "error making HTTP request")

// 			assert.Equal(t, tt.expectedCode, resp.StatusCode(), "Response code didn't match expected")
// 			// проверяем корректность полученного тела ответа, если мы его ожидаем
// 			// if tt.expectedBody != "" {
// 			// 	assert.JSONEq(t, tt.expectedBody, string(resp.Body()))
// 			// }
// 		})
// 	}
// }
