package tx

import "context"

//トランザクション境界。runner抽象
//ここでinfraimportしたら終わり。
//transaction境界はオーケストレーションに関する境界。domainに置かない
//usecaseは*sqlx.Txを引数に持ちたい(repoが使う。userrepo.New(q db.QueryerExecer))けどそれは依存おかしくなるから不可
//depsをusecaseに作ってそれをappでDIする
//[]型パラメータ（Generics）型は後から決める。動的な型
type Runner[D any] interface { //runnerは増やさない。
	WithinTx(ctx context.Context, fn func(ctx context.Context, deps D) error) error
}

//Depsパターン or UnitOfWorkパターン
