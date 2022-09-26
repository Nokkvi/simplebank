package api

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	db "github.com/nokkvi/simplebank/db/sqlc"
)

var account db.Account

func TestMain(m *testing.M) {
	account = randomAccount()
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}