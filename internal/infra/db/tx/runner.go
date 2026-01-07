package txrunner

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kou-etal/go_todo_app/internal/infra/db"
	//これするぐらいなら同じdbパッケージでrunner定義すればよくねと思うけどtxrunnerはアプリ都合やから分けたほうがいいんか
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
	//これは型チェックのためだけの抽象をimportやから可。実装をimportしてたらアウト。
)

//interface切りすぎるなってのは意味のないinterface。境界を守るためのinterfaceは必須
/*type UserRepo interface {
    Find(ctx context.Context, id ID) (*User, error)
}
をtype UserFinder interface {
    Find(ctx context.Context, id ID) (*User, error)
}
これは意味ない。
*/

type DepsFactory func(q db.QueryerExecer) usetx.Deps

//これdb.QueryerExecerを引数に持ってるからこれが引数強制してる
//解決策:deps作成をinfraで行う。でもそれしたらinfra責務大きくなる
//解決策:db.QueryerExecerじゃなくてusecaseでqueryer抽象作る。でもそれしたらusecaseがsqlxをimportしない限り
//namedexec使えないさらにせっかくsqlxwrapper作ってsqlx前提にしたのにusecaseでまたwrapper作るの良くわからん
//それらの折衷案の設計

type SQLxRunner struct {
	beginner   db.Beginner
	opts       *sql.TxOptions
	makeTxDeps DepsFactory
}

var _ usetx.Runner = (*SQLxRunner)(nil)

func New(beginner db.Beginner, opts *sql.TxOptions, makeTxDeps DepsFactory) *SQLxRunner {
	return &SQLxRunner{
		beginner:   beginner,
		opts:       opts,
		makeTxDeps: makeTxDeps,
	}
}

func (r *SQLxRunner) WithinTx(
	ctx context.Context,
	fn func(ctx context.Context, deps usetx.Deps) error,
) (retErr error) {
	//deferで追記するから名前付き戻り値
	tx, err := r.beginner.BeginTxx(ctx, r.opts) //tx作成
	if err != nil {
		return err
	}

	committed := false
	//想定できるならばerr。想定できない、コードのエラーはpanic
	//panicはスタックトレース付く。
	defer func() {
		//panicならばrollbackして再panic
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}

		//trueまで到達してない。
		if retErr != nil && !committed {
			//ロールバックすらできない
			if rbErr := tx.Rollback(); rbErr != nil {
				retErr = fmt.Errorf("tx rollback failed: %v (original: %w)", rbErr, retErr)
			}
		}
	}()

	deps := r.makeTxDeps(tx) //ここでtx(sqlx.Tx)代入できるのはrepositoryでnew引数をdbじゃなくてexecerqueryer抽象にしたから

	//usecaseで代入したやつを実行
	if err := fn(ctx, deps); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		//commit失敗はcloseされるからrollbackしなくていい。
		committed = true
		return err
	}
	committed = true
	return nil
}

//どっちが外側を意識する。usecase<-infra
//どこが具体実装でどこが抽象化を意識する。
//txrunnerはinfra実装、usecase抽象。
//普段はinfra実装、永続に関するdomain抽象やけど今回はトランザクション抽象やからdomain責務ではなくusecaseに抽象。これがだるい。
//そもそもusecase抽象の引数をどうするか。db.ExecerQueryer持たせたらそれは楽やけどinfra層に依存するからそれは不可。
//じゃあ当然普段repoからdb依存消す方式と同様にrepoのnewにdb周り含ませる。
//runnerはdepsかunitofworkどっち持たせるか。
//txをusecase structに持たせる設計かdeps持たせる設計の判断。
//txかどうかをappで分岐させたい場合は後者でもいいが今回はそうではないからtxrunnerをusecase structに持たせる。
//depsはinfraからのusecase importが業務の型にならないように抽象にする。
//factoryは業務ルールでもなんでもなく組み立てやからappかinfraに置く。depsはusecaseが欲しい契約(型)やからusecaseに置く。
