package emailverifyrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
)

/*
そういえばこれなんでポインタレシーバやっけ
値レシーバ　Repositoryも*Repositoryもメソッドを持つ
ポインタレシーバ　*Repositoryだけがそのメソッドを持つ
これから状態を持たせるかもしれない。
値レシーバはコピーしてもいい型
ポインタレシーバは共有されるオブジェクトという慣習
DIするstructはポインタレシーバ
*/
func (r *Repository) FindByTokenHashForUpdate(ctx context.Context, hash dverify.TokenHash) (*dverify.EmailVerificationToken, error) {
	//user_idだけでなくすべての行返す。usecaseがtoken.IsExpired(now)token.IsUsed()とかを使うからentity必須
	const q = `
SELECT
  id,
  user_id,
  token_hash,
  expires_at,
  used_at,
  created_at
FROM email_verification_tokens
WHERE token_hash = ?
FOR UPDATE;
`
	//for update。同じ行を別トランザクションが扱おうとするとロックする。同時にtokenhash検索して二回used_atを防ぐ。
	//使い切りトークンでは必須
	var rec EmailVerificationTokenRecord
	if err := r.q.GetContext(ctx, &rec, q, hash.Hash()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dverify.ErrNotFound
		} //定義されたエラーは吸収。そうでなければwrap
		//これエラー判定を別ファイルに分けたらここのdb依存なくなる。
		return nil, fmt.Errorf("emailverifyrepo find by token_hash for update: %w", err)
	}

	t, err := toEntity(rec)
	//これすでに関数側でerrorはwrapしてるけどreturn errか再びwrapかどっちがいいん
	//違う責務、境界が変わる、有益、ならばwrap。じゃあ吸収されるまではほとんどreturn errせずに再びラップ
	if err != nil {
		return nil, fmt.Errorf("emailverifyrepo map record to entity: %w", err)
	}

	return t, nil
}
