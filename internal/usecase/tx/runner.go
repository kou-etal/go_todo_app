package tx

import "context"

//トランザクション境界。runner抽象
//ここでinfraimportしたら終わり。
//transaction境界はオーケストレーションに関する境界。domainに置かない
//usecaseは*sqlx.Txを引数に持ちたい(repoが使う。userrepo.New(q db.QueryerExecer))けどそれは依存おかしくなるから不可
//depsをusecaseに作ってそれをappでDIする
type Runner interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, deps RegisterDeps) error) error //usecaseごとにtx
}

//Depsパターン or UnitOfWorkパターン
