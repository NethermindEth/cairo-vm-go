package zero

import (
	"testing"
)

func TestZeroHintUsort(t *testing.T) {
	t.Run("createUsortEnterScopeHinter", func(t *testing.T) {
		_, err := createUsortEnterScopeHinter(nil)
		if err != nil {
			t.Errorf("createUsortEnterScopeHinter() got err = %v", err)
			return
		}
	})
}
