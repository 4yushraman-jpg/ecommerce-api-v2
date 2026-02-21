package handlers

import (
	"bytes"
	"ecommerce-api-v2/internal/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterUserHandler_Success(t *testing.T) {
	db := setupTestDB()
	defer db.Close()

	handler := &UserHandler{DB: db}

	reqBody := models.RegisterUserRequest{
		Email:    "testuser@example.com",
		Password: "supersecretpassword",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.RegisterUserHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code 201, got %d", w.Code)
	}

	var count int
	err := db.QueryRow(req.Context(), "SELECT count(*) FROM users WHERE email = $1", "testuser@example.com").Scan(&count)

	if err != nil {
		t.Fatalf("Failed to query test database: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected exactly 1 user in the database, found %d", count)
	}
}

func TestRegisterUserHandler_DuplicateEmail(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	handler := &UserHandler{DB: db}

	reqBody := models.RegisterUserRequest{
		Email:    "duplicate@example.com",
		Password: "password123",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/users/register", bytes.NewReader(bodyBytes))
	w1 := httptest.NewRecorder()
	handler.RegisterUserHandler(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("Expected first user to be created (201), got %d", w1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/users/register", bytes.NewReader(bodyBytes))
	w2 := httptest.NewRecorder()
	handler.RegisterUserHandler(w2, req2)

	if w2.Code == http.StatusCreated {
		t.Errorf("Expected an error status code for duplicate email, but got 201 Created")
	}

	var count int
	err := db.QueryRow(req2.Context(), "SELECT count(*) FROM users WHERE email = $1", "duplicate@example.com").Scan(&count)

	if err != nil {
		t.Fatalf("Failed to query test database: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected exactly 1 user in the database, found %d", count)
	}
}

func TestLoginUserHandler_Success(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	handler := &UserHandler{DB: db}

	credentials := map[string]string{
		"email":    "logintest@example.com",
		"password": "securepassword123",
	}
	bodyBytes, _ := json.Marshal(credentials)

	reqReg := httptest.NewRequest(http.MethodPost, "/api/v1/users/register", bytes.NewReader(bodyBytes))
	wReg := httptest.NewRecorder()
	handler.RegisterUserHandler(wReg, reqReg)

	if wReg.Code != http.StatusCreated {
		t.Fatalf("Test setup failed: expected user to be created, got status %d", wReg.Code)
	}

	reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/users/login", bytes.NewReader(bodyBytes))
	wLogin := httptest.NewRecorder()

	handler.LoginUserHandler(wLogin, reqLogin)

	if wLogin.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK for valid login, got %d", wLogin.Code)
	}

	var responseBody map[string]interface{}
	if err := json.NewDecoder(wLogin.Body).Decode(&responseBody); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	token, exists := responseBody["token"]
	if !exists {
		t.Errorf("Expected 'token' in JSON response, but it was missing")
	}

	tokenStr, ok := token.(string)
	if !ok || tokenStr == "" {
		t.Errorf("Expected token to be a non-empty string")
	}
}

func TestLoginUserHandler_BadPassword(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	handler := &UserHandler{DB: db}

	credentials := map[string]string{
		"email":    "secureuser@example.com",
		"password": "correctpassword123",
	}
	bodyBytes, _ := json.Marshal(credentials)

	reqReg := httptest.NewRequest(http.MethodPost, "/api/v1/users/register", bytes.NewReader(bodyBytes))
	wReg := httptest.NewRecorder()
	handler.RegisterUserHandler(wReg, reqReg)

	badCredentials := map[string]string{
		"email":    "secureuser@example.com",
		"password": "wrongpassword_whoops",
	}
	badBodyBytes, _ := json.Marshal(badCredentials)

	reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/users/login", bytes.NewReader(badBodyBytes))
	wLogin := httptest.NewRecorder()

	handler.LoginUserHandler(wLogin, reqLogin)

	if wLogin.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 Unauthorized for bad password, got %d", wLogin.Code)
	}

	var responseBody map[string]interface{}
	json.NewDecoder(wLogin.Body).Decode(&responseBody)

	if _, exists := responseBody["token"]; exists {
		t.Errorf("CRITICAL SECURITY FAILURE: Token was returned despite a bad password!")
	}
}
