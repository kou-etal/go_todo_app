package userrepo

import (
	"context"
	"fmt"
	"time"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
)

func (r *Repository) MarkEmailVerified( //必要最低限の引数だけ受け取る。entity全部は受け取らない。それによって何を使うかもわかりやすい
	ctx context.Context,
	userID duser.UserID,
	verifiedAt time.Time,
) error {

	const q = `
UPDATE user
SET
  email_verified_at = ?,
  updated_at = ?,
  version = version + 1
WHERE
  id = ?
  AND email_verified_at IS NULL;
` //email_verified_at IS NULL;これはいい保険
	//わざわざnext_version作らずにこれが最適
	//n := verifiedAt.UTC().Truncate(time.Second)これはusecaseに責務寄せよう。

	res, err := r.q.ExecContext(
		ctx,
		q,
		verifiedAt,
		verifiedAt,
		userID.Value(),
	) //updatedatじゃなくてverifiedat(consume)に合わせてる。
	if err != nil {
		return fmt.Errorf("userrepo mark email verified execute: %w", err)
	}

	ra, err := res.RowsAffected() //何行成功したか
	if err != nil {
		return fmt.Errorf("userrepo mark email verified rowsaffected: %w", err)
	}

	if ra == 0 {
		//すでに更新されてる、そもそも存在しない
		//TODO:notfoundか楽観ロック分類するならばロジック記述
		return duser.ErrConflict
	}

	return nil
}
