package canonical

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

var pattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

func Validate(value *string) error {
	if value == nil {
		return nil
	}
	if !pattern.MatchString(*value) {
		return fmt.Errorf("canonical_id must be 1-128 characters using only letters, digits, '.', '_', or '-'")
	}
	if _, err := uuid.Parse(*value); err == nil {
		return fmt.Errorf("canonical_id must not be UUID-shaped")
	}
	return nil
}
