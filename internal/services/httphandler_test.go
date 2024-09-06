package services

//
//func TestHandler_Authenticated(t *testing.T) {
//	m := gomock.NewController(t)
//	defer m.Finish()
//
//	authMock := auth.NewMockAuth(m)
//	logger := mlog.NewMockLogger(m)
//
//	h := &handler{
//		logger: logger,
//		auth:   authMock,
//	}
//
//	// Test case: valid token
//	authMock.EXPECT().VerifyToken("valid-token").Return(&auth.JwtClaims{UserInfo: &auth.UserInfo{Name: "test-user"}}, true).Times(1)
//	req, _ := http.NewRequest("GET", "/", nil)
//	req.Header.Add("Authorization", "valid-token")
//	newReq, ok := h.authenticated(req)
//	assert.True(t, ok)
//	assert.Equal(t, "test-user", auth.MustGetUser(newReq.Context()).Name)
//
//	// Test case: invalid token
//	authMock.EXPECT().VerifyToken("invalid-token").Return(nil, false).Times(1)
//	req2, _ := http.NewRequest("GET", "/", nil)
//	req2.Header.Set("Authorization", "invalid-token")
//	_, ok = h.authenticated(req2)
//	assert.False(t, ok)
//}
//
//func TestHeaderMatcher(t *testing.T) {
//	// Test case: tracestate key
//	key, ok := headerMatcher("tracestate")
//	assert.True(t, ok)
//	assert.Equal(t, "tracestate", key)
//
//	// Test case: traceparent key
//	key, ok = headerMatcher("traceparent")
//	assert.True(t, ok)
//	assert.Equal(t, "traceparent", key)
//
//	// Test case: other key
//	key, ok = headerMatcher("other")
//	assert.False(t, ok)
//	assert.Equal(t, "", key)
//
//	// Test case: empty key
//	key, ok = headerMatcher("")
//	assert.False(t, ok)
//	assert.Equal(t, "", key)
//}
//
//func Test_handleBinaryFileUpload(t *testing.T) {
//	m := gomock.NewController(t)
//	defer m.Finish()
//	db, _ := data.NewSqliteDB()
//	defer db.Close()
//	req := &http.Request{
//		Form: map[string][]string{},
//	}
//	up := uploader.NewMockUploader(m)
//	h := &handler{
//		logger:   mlog.NewForConfig(nil),
//		auth:     auth.NewMockAuth(m),
//		uploader: up,
//		data:     data.NewDataImpl(&data.NewDataParams{DB: db}),
//	}
//
//	rr := httptest.NewRecorder()
//	h.handleBinaryFileUpload(rr, req)
//	assert.Equal(t, 400, rr.Code)
//
//	postData :=
//		`value2
//--xxx
//Content-Disposition: form-data; name="file"; filename="a.txt"
//Content-Type: application/octet-stream
//Content-Transfer-Encoding: binary
//
//binary data
//--xxx--
//	`
//	req2 := &http.Request{
//		Method: "POST",
//		Header: http.Header{"Content-Type": {`multipart/form-data; boundary=xxx`}},
//		Body:   io.NopCloser(strings.NewReader(postData)),
//	}
//
//	req2.Form = make(url.Values)
//	req2 = req2.WithContext(auth.SetUser(req2.Context(), &auth.UserInfo{Name: "duc"}))
//	rr2 := httptest.NewRecorder()
//
//	up.EXPECT().Type().Return(schematype.Local)
//	up.EXPECT().Disk("users").Return(up)
//	finfo := uploader.NewMockFileInfo(m)
//	up.EXPECT().Put(gomock.Any(), gomock.Any()).Return(finfo, nil)
//	finfo.EXPECT().Path().Return("/app.txt")
//	finfo.EXPECT().Size().Return(uint64(1000))
//	h.handleBinaryFileUpload(rr2, req2)
//	assert.Equal(t, 201, rr2.Code)
//	f, _ := db.File.Query().First(context.TODO())
//	assert.Equal(t, "application/json", rr2.Header().Get("Content-Type"))
//	assert.Equal(t, "/app.txt", f.Path)
//	assert.Equal(t, uint64(1000), f.Size)
//	assert.Equal(t, "duc", f.Username)
//}
//
//func Test_download(t *testing.T) {
//	m := gomock.NewController(t)
//	defer m.Finish()
//	db, _ := data.NewSqliteDB()
//	defer db.Close()
//	h := &handler{
//		logger: mlog.NewForConfig(nil),
//	}
//
//	recorder := httptest.NewRecorder()
//	h.download(recorder, "f.txt", strings.NewReader("aaa"))
//	assert.Equal(t, "application/octet-stream", recorder.Header().Get("Content-Type"))
//	assert.Equal(t, fmt.Sprintf(`attachment; filename="%s"`, url.QueryEscape("f.txt")), recorder.Header().Get("Content-Disposition"))
//	assert.Equal(t, "0", recorder.Header().Get("Expires"))
//	assert.Equal(t, "binary", recorder.Header().Get("Content-Transfer-Encoding"))
//	assert.Equal(t, "*", recorder.Header().Get("Access-Control-Expose-Headers"))
//	assert.Equal(t, "aaa", recorder.Body.String())
//	assert.Equal(t, http.StatusOK, recorder.Code)
//}
