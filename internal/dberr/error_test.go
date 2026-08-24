package dberr

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestIsDuplicateKeyError(t *testing.T) {
	duplicate := &mysql.MySQLError{Number: 1062, Message: "duplicate"}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "mysql duplicate", err: duplicate, want: true},
		{name: "wrapped duplicate", err: fmt.Errorf("create account: %w", duplicate), want: true},
		{name: "other mysql error", err: &mysql.MySQLError{Number: 1045}, want: false},
		{name: "ordinary error", err: errors.New("boom"), want: false},
		{name: "nil", err: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDuplicateKeyError(tt.err); got != tt.want {
				t.Fatalf("IsDuplicateKeyError() = %t, want %t", got, tt.want)
			}
		})
	}
}
