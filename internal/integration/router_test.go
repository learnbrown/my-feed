package integration

import (
	dbtest "my_feed/internal/db/testutil"
	"my_feed/internal/router"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	sqlDB := dbtest.SetupTestDB(t)
	dbtest.CleanTestDB(t, sqlDB)

	return router.SetRouter(sqlDB)
}
