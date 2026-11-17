package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"app/internal/config"
	"app/internal/database"
)

func TestSetupRouter_PureApi(t *testing.T) {
	cfg := &config.Config{
		AppEnv: "production",
		AppKey: "test-secret",
	}

	db, err := database.ConnectTest()
	assert.NoError(t, err)

	router := SetupRouter(cfg, db)

	// 1. Test 404 unmatched route
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/random-unmatched-route", nil)
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusNotFound, w1.Code)
	assert.Contains(t, w1.Body.String(), "Endpoint not found")

	// 2. Test 404 unmatched API route
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/nonexistent", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNotFound, w2.Code)
	assert.Contains(t, w2.Body.String(), "Endpoint not found")

	// 3. Test Public templates endpoint
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/api/templates", nil)
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)

	// 4. Test Swagger UI endpoint
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	req4.RequestURI = "/swagger/index.html"
	router.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)
	assert.Contains(t, w4.Body.String(), "swagger-ui")

	// 5. Test Swagger Initializer points to api.json
	w5 := httptest.NewRecorder()
	req5, _ := http.NewRequest(http.MethodGet, "/swagger/swagger-initializer.js", nil)
	req5.RequestURI = "/swagger/swagger-initializer.js"
	router.ServeHTTP(w5, req5)
	assert.Equal(t, http.StatusOK, w5.Code)
	assert.Contains(t, w5.Body.String(), `url: "/swagger/api.json"`)

	// 6. Test Swagger api.json endpoint
	w6 := httptest.NewRecorder()
	req6, _ := http.NewRequest(http.MethodGet, "/swagger/api.json", nil)
	req6.RequestURI = "/swagger/api.json"
	router.ServeHTTP(w6, req6)
	assert.Equal(t, http.StatusOK, w6.Code)
	assert.Contains(t, w6.Body.String(), `"swagger": "2.0"`)
	assert.Contains(t, w6.Body.String(), "Delcom Todolist REST API")

	// 7. Test Swagger doc.json backward compatibility endpoint
	w7 := httptest.NewRecorder()
	req7, _ := http.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	req7.RequestURI = "/swagger/doc.json"
	router.ServeHTTP(w7, req7)
	assert.Equal(t, http.StatusOK, w7.Code)
	assert.Contains(t, w7.Body.String(), `"swagger": "2.0"`)

	// 8. Test Root api.json endpoint
	w8 := httptest.NewRecorder()
	req8, _ := http.NewRequest(http.MethodGet, "/api.json", nil)
	router.ServeHTTP(w8, req8)
	assert.Equal(t, http.StatusOK, w8.Code)
	assert.Contains(t, w8.Body.String(), `"swagger": "2.0"`)
}
