package userrepo

import "time"

//これ別パッケージから使わないしフィールド大文字やめてカプセルにしたほうがよくね
//これ大文字じゃないとdbマッピングできないから不可
//そもそもdomainはメソッドやルールを持ってるから不変条件を守るためにカプセルにした
//UserRecordはただのDTO。ルールも持たない。
type UserRecord struct {
	ID              string     `db:"id"`
	Email           string     `db:"email"`
	PasswordHash    string     `db:"password_hash"`
	UserName        string     `db:"user_name"`
	EmailVerifiedAt *time.Time `db:"email_verified_at"`
	Created         time.Time  `db:"created_at"`
	Updated         time.Time  `db:"updated_at"`
	Version         uint64     `db:"version"`
}

//EmailVerifiedAt *time.Time `db:"email_verified_at"`これいまGo nil->sql nullが自動変換されてる状況
//join(テーブルくっつけて出力)
//left join(あるデータは必ず返す)でnullで埋められた際それが自動変換によるnullかjoinのnullかわからない。
//その他自動変換につきdriverの仕様によるという問題ある
//TODO:だから実務ではEmailVerifiedAt sql.NullTimeで確実にnullを扱うことが多い。
