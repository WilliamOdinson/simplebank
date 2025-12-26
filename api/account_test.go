package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/WilliamOdinson/simplebank/db/mock"
	db "github.com/WilliamOdinson/simplebank/db/sqlc"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/lib/pq"
	"go.uber.org/mock/gomock"
)

func TestGetAccountAPI(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
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
			name:      "NotFound",
			accountID: account.ID,
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
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateAccountAPI(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		body          map[string]any
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]any{
				"owner":    account.Owner,
				"currency": "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), db.CreateAccountParams{
						Owner:    account.Owner,
						Currency: "USD",
						Balance:  0,
					}).
					Times(1).
					Return(db.Account{
						ID:       account.ID,
						Owner:    account.Owner,
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
			name: "InternalError",
			body: map[string]any{
				"owner":    account.Owner,
				"currency": "USD",
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
				"owner":    account.Owner,
				"currency": "INVALID",
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
			name: "MissingOwner",
			body: map[string]any{
				"currency": "USD",
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
			body: map[string]any{
				"owner": account.Owner,
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
				"owner":    account.Owner,
				"currency": "EUR",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), db.CreateAccountParams{
						Owner:    account.Owner,
						Currency: "EUR",
						Balance:  0,
					}).
					Times(1).
					Return(db.Account{
						ID:       account.ID,
						Owner:    account.Owner,
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
				"owner":    account.Owner,
				"currency": "CNY",
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
				"owner":    "nonexistent_user",
				"currency": "USD",
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
				"owner":    account.Owner,
				"currency": "USD",
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
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccountsAPI(t *testing.T) {
	owner := gofakeit.LetterN(10)
	n := 5
	accounts := make([]db.Account, n)
	for i := 0; i < n; i++ {
		accounts[i] = db.Account{
			ID:       gofakeit.Int64(),
			Owner:    owner,
			Balance:  int64(gofakeit.Price(0, 10000)),
			Currency: "USD",
		}
	}

	testCases := []struct {
		name          string
		owner         string
		pageID        int32
		pageSize      int32
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			owner:    owner,
			pageID:   1,
			pageSize: 5,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), db.ListAccountsParams{
						Owner:  owner,
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
			name:     "SecondPage",
			owner:    owner,
			pageID:   2,
			pageSize: 5,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), db.ListAccountsParams{
						Owner:  owner,
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
			owner:    owner,
			pageID:   1,
			pageSize: 10,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), db.ListAccountsParams{
						Owner:  owner,
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
			owner:    owner,
			pageID:   1,
			pageSize: 5,
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
			owner:    owner,
			pageID:   0,
			pageSize: 5,
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
			owner:    owner,
			pageID:   -1,
			pageSize: 5,
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
			owner:    owner,
			pageID:   1,
			pageSize: 4,
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
			owner:    owner,
			pageID:   1,
			pageSize: 11,
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
			url := fmt.Sprintf("/accounts?owner=%s&page_id=%d&page_size=%d", tc.owner, tc.pageID, tc.pageSize)
			request := httptest.NewRequest(http.MethodGet, url, nil)

			// 5. Serve the request
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccountsAPI_MissingParams(t *testing.T) {
	owner := gofakeit.LetterN(10)

	testCases := []struct {
		name          string
		url           string
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "MissingOwner",
			url:  "/accounts?page_id=1&page_size=5",
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name: "MissingPageID",
			url:  fmt.Sprintf("/accounts?owner=%s&page_size=5", owner),
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf("expected status code 400, got %d", recorder.Code)
				}
			},
		},
		{
			name: "MissingPageSize",
			url:  fmt.Sprintf("/accounts?owner=%s&page_id=1", owner),
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
