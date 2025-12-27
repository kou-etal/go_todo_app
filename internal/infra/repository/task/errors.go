package taskrepo

import (
	"database/sql"
	"errors"
)

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

//TODO:エラー伝搬適当すぎる。それぞれ定義してusecaseで吸収か、domainに寄せるか統一しよう
