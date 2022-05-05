// nolint: testpackage
package repository

import (
	"testing"
)

func TestUser_GetEmail(t *testing.T) {
	t.Parallel()
	t.Run("success()", func(t *testing.T) {
		t.Parallel()
		const expect = "test"
		u := &User{
			email: expect,
		}
		if got := u.GetEmail(); got != expect {
			t.Errorf("User.GetEmail() = %v, want %v", got, expect)
		}
	})
}
