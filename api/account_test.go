package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/WilliamOdinson/simplebank/db/mock"
	db "github.com/WilliamOdinson/simplebank/db/sqlc"
	"github.com/brianvoe/gofakeit/v7"
	"go.uber.org/mock/gomock"
)

func TestGetAccountAPI(t *testing.T) {
	account := randomAccount()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)

	store.EXPECT().
		GetAccount(gomock.Any(), gomock.Eq(account.ID)). // expect to get account with specific ID under any context
		Times(1).                                        // expect to be called once
		Return(account, nil)                             // return the account and no error

	// start test server and send request
	server := NewServer(store)
	recorder := httptest.NewRecorder()

	url := fmt.Sprintf("/accounts/%d", account.ID)
	request := httptest.NewRequest(http.MethodGet, url, nil)
	server.router.ServeHTTP(recorder, request)

	// check response
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status code 200, got %d", recorder.Code)
	}

	// check response body
	var gotAccount db.Account
	if err := json.NewDecoder(recorder.Body).Decode(&gotAccount); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if gotAccount != account {
		t.Errorf("expected account %+v, got %+v", account, gotAccount)
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
