package refreshrepo

import (
	"database/sql"
	"errors"
)

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
