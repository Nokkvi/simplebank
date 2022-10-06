package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/nokkvi/simplebank/db/sqlc"
	"github.com/nokkvi/simplebank/token"
)

type GetEntryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getEntry(ctx *gin.Context) {
	var req GetEntryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	entry, err := server.store.GetEntry(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	
	account, err := server.store.GetAccount(ctx, entry.AccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	
	authPayload := ctx.MustGet(autorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		err := errors.New("Account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, entry)
}

type ListEntriesRequest struct {
	AccountID int64 `form:"account_id" binding:"required,min=1"`
	PageID     int32 `form:"page_id" binding:"required,min=1"`
	PageSize    int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listEntries(ctx *gin.Context) {
	var req ListEntriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccount(ctx, req.AccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	
	authPayload := ctx.MustGet(autorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		err := errors.New("Account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.ListEntriesParams{
		AccountID: req.AccountID,
		Limit: req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}
	entries, err := server.store.ListEntries(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, entries)
}