package userrepo

import (
	"fmt"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
)

// これ別に関数公開する必要ないからRecordToEntityではなく
// ポインタを引数にしたらコピー減って軽くなるっていうけどそれはUserRecordレベルやったら誤差。
// 逆にr *UserRecordこれはnilを与えた時にpanicになるっていうデメリットある。
func toEntity(r UserRecord) (*duser.User, error) {

	//	id := duser.UserID(r.ID)。これDBを信用するって意味では別にいいけど
	//せっかくparse作ったから使おう
	id, err := duser.ParseUserID(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user record id=%s field=id: %w", r.ID, err)
	}
	//ここのnewはvalidation+型変換のうちただの型からdomainの型に変換する意味。usecaseから使うnewとは意味合い違うけど再利用。
	//TODO:でもよく考えたらもしvoに新たな制約つくったら既存のデータ復元できなくなる。ゆえに全部にreconstruct作るのが丸いかも
	//エラーはhandlerで公開していいかを判定するイメージ
	email, err := duser.NewUserEmail(r.Email)
	if err != nil {
		//emailとかpasswordは当然エラーに含めない。
		return nil, fmt.Errorf("invalid user record id=%s field=email: %w", r.ID, err)
	} //これはDBがバグってることによるエラーゆえにdomainのエラーは使わずにwrapして返す
	//passwordはusecaseから使うnewと完全にわけてrepoから使う用のreconstruct定義。
	password, err := duser.ReconstructUserPassword(r.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("invalid user record id=%s field=password: %w", r.ID, err)
	}
	name, err := duser.NewUserName(r.UserName)
	if err != nil {
		return nil, fmt.Errorf("invalid user record id=%s field=userName: %w", r.ID, err)
	}

	/*ReconstructUserを作らなければuser := &User{
	      id: id,
	      email: email,
	      ...
	  	となりrepoがdomainに関与しすぎる
	*/

	return duser.ReconstructUser(
		id,
		email,
		password,
		name,
		r.EmailVerifiedAt,
		r.Created,
		r.Updated,
		r.Version,
	), nil
}

//これもuserrrecordをポインタやめる。
//u *duser.Userこれはルール持ってるからポインタ
// これ別に関数公開する必要ないからEntityToRecordではなく

func toRecord(u *duser.User) UserRecord {

	return UserRecord{
		ID: u.ID().Value(),
		//Email:       string(u.Email().Value()),これstringは保険にしてもいらん
		Email:           u.Email().Value(),
		PasswordHash:    u.Password().Hash(), //hasherの段階でhashをstringで返してる。
		UserName:        u.UserName().Value(),
		EmailVerifiedAt: u.EmailVerifiedAt(),
		Created:         u.CreatedAt(),
		//ここはnormalizationしたらデータが変わるからrepoを信用する。
		Updated: u.UpdatedAt(),
		Version: u.Version(),
	}
}
