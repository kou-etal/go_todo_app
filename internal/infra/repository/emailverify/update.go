package emailverifyrepo

import (
	"context"
	"fmt"

	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
)

func (r *Repository) Update(
	ctx context.Context,
	t *dverify.EmailVerificationToken,
) error {
	//repoはnowを持たない。usecaseでconsumeする

	const q = `
UPDATE email_verification_tokens
SET
  used_at = :used_at
WHERE
  id = :id
  AND used_at IS NULL;
`
	//ここ別にAND used_at IS NULL、AND expires_at > now;こうやってすることもできるけどこれはrepoの責務大きくなる。テストしにくい。
	//ゆえにdomainに責務持たせる。
	//FOR UPDATEはSELECT文で使う。ここでは使わない。updateはimplicit lock。
	//select(for update)->有効性確認(usecase)->update->commitのイメージ
	//AND used_at IS NULLこれで同時回避
	rec := toRecord(t)

	res, err := r.q.NamedExecContext(ctx, q, rec)
	if err != nil {
		return fmt.Errorf("emailverifyrepo consume execute: %w", err)
	}
	//さらに同時回避
	ra, err := res.RowsAffected()
	//TODO:ここはちゃんとconflictとnotfound分けるべき
	if err != nil {
		return fmt.Errorf("emailverifyrepo consume rowsaffected: %w", err)
	}

	if ra == 0 {
		return dverify.ErrConflict
	}

	return nil
}
