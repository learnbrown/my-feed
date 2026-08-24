package integration

import (
	"errors"
	"testing"

	"my_feed/internal/account"
	dbtest "my_feed/internal/db/testutil"

	"gorm.io/gorm"
)

func TestRegisterDuplicateAndConcurrentRequestsWithDB(t *testing.T) {
	sqlDB := dbtest.SetupTestDB(t)
	dbtest.CleanTestDB(t, sqlDB)

	service := account.NewAccountService(account.NewAccountRepo(sqlDB), nil)
	if _, err := service.Register("duplicate-user", "secret123"); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	if _, err := service.Register("duplicate-user", "secret123"); !errors.Is(err, account.ErrUsernameExists) {
		t.Fatalf("duplicate Register() error = %v, want ErrUsernameExists", err)
	}
	assertAccountRows(t, sqlDB, "duplicate-user", 1)

	const concurrency = 8
	results := make(chan error, concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			_, err := service.Register("concurrent-user", "secret123")
			results <- err
		}()
	}

	var successes int
	for i := 0; i < concurrency; i++ {
		err := <-results
		switch {
		case err == nil:
			successes++
		case errors.Is(err, account.ErrUsernameExists):
			// Expected for every request that loses the unique-index race.
		default:
			t.Errorf("concurrent Register() unexpected error = %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("successful concurrent registrations = %d, want 1", successes)
	}
	assertAccountRows(t, sqlDB, "concurrent-user", 1)
}

func assertAccountRows(t *testing.T, sqlDB *gorm.DB, username string, want int64) {
	t.Helper()

	var count int64
	if err := sqlDB.Model(&account.Account{}).Where("username = ?", username).Count(&count).Error; err != nil {
		t.Fatalf("count account %q: %v", username, err)
	}
	if count != want {
		t.Fatalf("account rows for %q = %d, want %d", username, count, want)
	}
}
