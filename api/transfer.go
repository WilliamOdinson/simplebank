package api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/WilliamOdinson/simplebank/db/sqlc"
	"github.com/WilliamOdinson/simplebank/token"
	"github.com/gin-gonic/gin"
)

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1,nefield=FromAccountID"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Check if from and to account is valid
	fromAccount, valid := server.validAccount(ctx, req.FromAccountID, req.Currency)
	if !valid {
		err := fmt.Errorf("from account %d is not valid", req.FromAccountID)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if fromAccount.Owner != authPayload.Username {
		err := fmt.Errorf("from account %d does not belong to the authenticated user", req.FromAccountID)
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	if _, valid := server.validAccount(ctx, req.ToAccountID, req.Currency); !valid {
		err := fmt.Errorf("to account %d is not valid", req.ToAccountID)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// validAccount checks if the account with given ID exists and if its currency matches the provided one.
func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err == sql.ErrNoRows {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return account, false
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}

	if account.Currency != currency {
		err := fmt.Errorf("account ID %d currency mismatch: expected %s, got %s", accountID, currency, account.Currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}

	return account, true
}
