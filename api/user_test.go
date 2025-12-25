package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/WilliamOdinson/simplebank/db/mock"
	db "github.com/WilliamOdinson/simplebank/db/sqlc"
	"github.com/WilliamOdinson/simplebank/util"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/lib/pq"
	"go.uber.org/mock/gomock"
)

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          map[string]any
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]any{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusOK {
					t.Errorf("expected status code 200, got %d", recorder.Code)
				}
			},
		},
		{
			name: "InternalError",
			body: map[string]any{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, fmt.Errorf("internal error"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusInternalServerError {
					t.Errorf("expected status code 500, got %d", recorder.Code)
				}
			},
		},
		{
			name: "DuplicateUsername",
			body: map[string]any{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusForbidden {
					t.Errorf("expected status code 403, got %d", recorder.Code)
				}
			},
		},
		{
			name: "MissingUsername",
			body: map[string]any{
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name: "MissingPassword",
			body: map[string]any{
				"username":  user.Username,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name: "MissingFullName",
			body: map[string]any{
				"username": user.Username,
				"password": password,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name: "MissingEmail",
			body: map[string]any{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name: "InvalidEmail",
			body: map[string]any{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     "invalid-email",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name: "PasswordTooShort",
			body: map[string]any{
				"username":  user.Username,
				"password":  "short",
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name: "InvalidUsername",
			body: map[string]any{
				"username":  "user@name",
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name: "InvalidJSON",
			body: nil,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			var body []byte
			if tc.body != nil {
				body, _ = json.Marshal(tc.body)
			} else {
				body = []byte("invalid json")
			}

			request := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
			request.Header.Set("Content-Type", "application/json")

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomUser(t *testing.T) (db.User, string) {
	password := gofakeit.Password(true, true, true, false, false, 16)
	hashedPassword, err := util.HashPassword(password)
	if err != nil {
		t.Fatal("Cannot hash password:", err)
	}

	user := db.User{
		Username:       gofakeit.Username(),
		HashedPassword: hashedPassword,
		FullName:       gofakeit.Name(),
		Email:          gofakeit.Email(),
	}

	return user, password
}
