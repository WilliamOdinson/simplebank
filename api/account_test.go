package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/WilliamOdinson/simplebank/db/mock"
	db "github.com/WilliamOdinson/simplebank/db/sqlc"
	"github.com/WilliamOdinson/simplebank/token"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/lib/pq"
	"go.uber.org/mock/gomock"
)

func TestGetAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccountForUser(user.Username)

	testCases := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusOK {
					t.Errorf("expected status code 200, got %d", recorder.Code)
				}
				var gotAccount db.Account
				if err := json.NewDecoder(recorder.Body).Decode(&gotAccount); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				if gotAccount != account {
					t.Errorf("expected account %+v, got %+v", account, gotAccount)
				}
			},
		},
		{
			name:      "UnauthorizedUser",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "other_user", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusUnauthorized {
					t.Errorf("expected status code 401, got %d", recorder.Code)
				}
			},
		},
		{
			name:      "NoAuthorization",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// Don't add authorization
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusUnauthorized {
					t.Errorf("expected status code 401, got %d", recorder.Code)
				}
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows) // Simulate not found error
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusNotFound {
					t.Errorf("expected status code 404, got %d", recorder.Code)
				}
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, fmt.Errorf("internal error"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusInternalServerError {
					t.Errorf("expected status code 500, got %d", recorder.Code)
				}
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name:      "NegativeID",
			accountID: -1,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
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
			// 1. Setup controller and mock store
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 2. Build stubs
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// 3. Create server and recorder
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// 4. Create request
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request := httptest.NewRequest(http.MethodGet, url, nil)

			// 5. Serve the request
			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccountForUser(user.Username)

	testCases := []struct {
		name          string
		body          map[string]any
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]any{
				"currency": "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), db.CreateAccountParams{
						Owner:    user.Username,
						Currency: "USD",
						Balance:  0,
					}).
					Times(1).
					Return(db.Account{
						ID:       account.ID,
						Owner:    user.Username,
						Currency: "USD",
						Balance:  0,
					}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusOK {
					t.Errorf("expected status code 200, got %d", recorder.Code)
				}
			},
		},
		{
			name: "NoAuthorization",
			body: map[string]any{
				"currency": "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// Don't add authorization
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusUnauthorized {
					t.Errorf("expected status code 401, got %d", recorder.Code)
				}
			},
		},
		{
			name: "InternalError",
			body: map[string]any{
				"currency": "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, fmt.Errorf("internal error"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusInternalServerError {
					t.Errorf("expected status code 500, got %d", recorder.Code)
				}
			},
		},
		{
			name: "InvalidCurrency",
			body: map[string]any{
				"currency": "INVALID",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name: "MissingCurrency",
			body: map[string]any{},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name: "EUR_Currency",
			body: map[string]any{
				"currency": "EUR",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), db.CreateAccountParams{
						Owner:    user.Username,
						Currency: "EUR",
						Balance:  0,
					}).
					Times(1).
					Return(db.Account{
						ID:       account.ID,
						Owner:    user.Username,
						Currency: "EUR",
						Balance:  0,
					}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusOK {
					t.Errorf("expected status code 200, got %d", recorder.Code)
				}
			},
		},
		{
			name: "UnsupportedCurrency_CNY",
			body: map[string]any{
				"currency": "CNY",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name: "ForeignKeyViolation",
			body: map[string]any{
				"currency": "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, &pq.Error{Code: "23503"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusForbidden {
					t.Errorf("expected status code 403, got %d", recorder.Code)
				}
			},
		},
		{
			name: "UniqueViolation",
			body: map[string]any{
				"currency": "USD",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusForbidden {
					t.Errorf("expected status code 403, got %d", recorder.Code)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Setup controller and mock store
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 2. Build stubs
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// 3. Create server and recorder
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// 4. Create request body
			var body []byte
			if tc.body != nil {
				body, _ = json.Marshal(tc.body)
			} else {
				body = []byte("invalid json")
			}
			request := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body))
			request.Header.Set("Content-Type", "application/json")

			// 5. Serve the request
			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccountsAPI(t *testing.T) {
	user, _ := randomUser(t)
	n := 5
	accounts := make([]db.Account, n)
	for i := 0; i < n; i++ {
		accounts[i] = randomAccountForUser(user.Username)
	}

	testCases := []struct {
		name          string
		pageID        int32
		pageSize      int32
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			pageID:   1,
			pageSize: 5,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), db.ListAccountsParams{
						Owner:  user.Username,
						Limit:  5,
						Offset: 0,
					}).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusOK {
					t.Errorf("expected status code 200, got %d", recorder.Code)
				}
			},
		},
		{
			name:     "NoAuthorization",
			pageID:   1,
			pageSize: 5,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// Don't add authorization
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusUnauthorized {
					t.Errorf("expected status code 401, got %d", recorder.Code)
				}
			},
		},
		{
			name:     "SecondPage",
			pageID:   2,
			pageSize: 5,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), db.ListAccountsParams{
						Owner:  user.Username,
						Limit:  5,
						Offset: 5,
					}).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusOK {
					t.Errorf("expected status code 200, got %d", recorder.Code)
				}
			},
		},
		{
			name:     "MaxPageSize",
			pageID:   1,
			pageSize: 10,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), db.ListAccountsParams{
						Owner:  user.Username,
						Limit:  10,
						Offset: 0,
					}).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusOK {
					t.Errorf("expected status code 200, got %d", recorder.Code)
				}
			},
		},
		{
			name:     "InternalError",
			pageID:   1,
			pageSize: 5,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, fmt.Errorf("internal error"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusInternalServerError {
					t.Errorf("expected status code 500, got %d", recorder.Code)
				}
			},
		},
		{
			name:     "InvalidPageID",
			pageID:   0,
			pageSize: 5,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name:     "NegativePageID",
			pageID:   -1,
			pageSize: 5,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name:     "PageSizeTooSmall",
			pageID:   1,
			pageSize: 4,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name:     "PageSizeTooLarge",
			pageID:   1,
			pageSize: 11,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
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
			// 1. Setup controller and mock store
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 2. Build stubs
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// 3. Create server and recorder
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// 4. Create request
			url := fmt.Sprintf("/accounts?page_id=%d&page_size=%d", tc.pageID, tc.pageSize)
			request := httptest.NewRequest(http.MethodGet, url, nil)

			// 5. Serve the request
			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccountsAPI_MissingParams(t *testing.T) {
	user, _ := randomUser(t)

	testCases := []struct {
		name          string
		url           string
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "MissingPageID",
			url:  "/accounts?page_size=5",
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name: "MissingPageSize",
			url:  "/accounts?page_id=1",
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name: "MissingAllParams",
			url:  "/accounts",
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Setup controller and mock store
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 2. Create mock store
			store := mockdb.NewMockStore(ctrl)
			store.EXPECT().
				ListAccounts(gomock.Any(), gomock.Any()).
				Times(0)

			// 3. Create server and recorder
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// 4. Create request
			request := httptest.NewRequest(http.MethodGet, tc.url, nil)
			addAuthorization(t, request, server.tokenMaker, authorizationTypeBearer, user.Username, time.Minute)

			// 5. Serve the request
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestServerStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	server := newTestServer(t, store)

	// Test that Start returns an error for invalid address
	err := server.Start("invalid-address-that-will-fail")
	if err == nil {
		t.Error("expected error for invalid address, got nil")
	}
}

func randomAccount() db.Account {
	return db.Account{
		ID:       gofakeit.Int64(),
		Owner:    gofakeit.Name(),
		Balance:  int64(gofakeit.Price(0, 10000)),
		Currency: gofakeit.CurrencyShort(),
	}
}

func randomAccountForUser(owner string) db.Account {
	return db.Account{
		ID:       gofakeit.Int64(),
		Owner:    owner,
		Balance:  int64(gofakeit.Price(0, 10000)),
		Currency: "USD",
	}
}
