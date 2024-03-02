package main

//type GetAllNodesSuite struct {
//	suite.Suite
//
//	// db is the database used for testing.
//	db *dataaccess.MockDb
//}
//
//func TestIndexSuite(t *testing.T) {
//	suite.Run(t, new(GetAllNodesSuite))
//}
//
//func (s *GetAllNodesSuite) SetupTest() {
//	s.db = new(dataaccess.MockDb)
//	dataaccess.DB = s.db
//}
//
//func (s *GetAllNodesSuite) TearDownTest() {
//	s.db.AssertExpectations(s.T())
//	s.db = nil
//}
//
//func (s *GetAllNodesSuite) TestIndexHandlerAPI() {
//	m := s.db
//
//	now := time.Now().UTC()
//
//	runs := []*entities.PuppetRun{
//		{
//			Fqdn:     "test1",
//			Env:      summary.Environment_PRODUCTION,
//			ExecTime: entities.Datetime(now),
//			Runtime:  entities.Duration(10 * time.Second),
//		},
//		{
//			Fqdn:     "test2",
//			Env:      summary.Environment_STAGING,
//			ExecTime: entities.Datetime(now),
//			Runtime:  entities.Duration(10 * time.Second),
//		},
//		{
//			Fqdn:     "test3",
//			Env:      summary.Environment_PRODUCTION,
//			ExecTime: entities.Datetime(now),
//			Runtime:  entities.Duration(10 * time.Second),
//		},
//	}
//
//	// Test the index handler with the API.
//	m.On("GetRuns", mock.AnythingOfType("context.backgroundCtx")).Return(runs, nil).Once()
//	m.On("GetHistory", mock.AnythingOfType("context.backgroundCtx"), summary.Environment_PRODUCTION, summary.Environment_STAGING, summary.Environment_DEVELOPMENT).Return([]*entities.PuppetHistory{}, nil).Once()
//	m.On("GetEnvironments", mock.AnythingOfType("context.backgroundCtx")).Return([]summary.Environment{summary.Environment_PRODUCTION, summary.Environment_STAGING, summary.Environment_DEVELOPMENT}, nil).Once()
//
//	w := httptest.NewRecorder()
//	r := httptest.NewRequest("GET", "/api/nodes", nil)
//
//	svc := new(webService)
//
//	svc.GetAllNodes(w, r)
//
//	s.Equal(200, w.Code)
//	s.Equal("application/json", w.Header().Get("Content-Type"))
//
//	// Compare the response.
//	expected := "[{\"fqdn\":\"test1\",\"env\":\"PRODUCTION\",\"state\":\"\",\"exec_time\":\"" + now.Format(time.RFC3339) + "\",\"runtime\":\"10s\"},{\"fqdn\":\"test2\",\"env\":\"STAGING\",\"state\":\"\",\"exec_time\":\"" + now.Format(time.RFC3339) + "\",\"runtime\":\"10s\"},{\"fqdn\":\"test3\",\"env\":\"PRODUCTION\",\"state\":\"\",\"exec_time\":\"" + now.Format(time.RFC3339) + "\",\"runtime\":\"10s\"}]\n"
//	s.Require().Equal(expected, w.Body.String())
//}
