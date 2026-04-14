//auth周りのヘルパー。
import http from "k6/http";
import { BASE_URL, SEED_PASSWORD, SEED_PREFIX, SEED_USERS } from "./config.js";

//桁数
const PAD_WIDTH = Math.max(3, String(SEED_USERS).length);

export function login(userNum) {
  const digits = String(userNum).padStart(PAD_WIDTH, "0");
  //TODO:これ命名良くない。
  const email = `${SEED_PREFIX}-${digits}@example.com`;

  const res = http.post(//k6パターン。
    `${BASE_URL}/users/login`,//baseurlはnormalize済み
    JSON.stringify({ email, password: SEED_PASSWORD }),
    { headers: { "Content-Type": "application/json" }, tags: { name: "POST /users/login" } },
    //タグがないとURLで分類。その場合grafana側でだるい。ここはURL一定やからなくてもいいけどPATCHで使うからすべてにタグつける。
  );//JSON.stringify必須。body にオブジェクトを直接与えると form-urlencoded として送られる。
  // ゆえにJSON API なら必ず stringify する

  if (res.status !== 200) {
    console.error(`login failed for ${email}: ${res.status} ${res.body}`);
    return null;//一つの失敗ですべて止めないためにnull返す。
  }

  const body = res.json();
  //これ200やけどおかしい場合はerrorでるからcatchしてもいいがsmokeテストでAPIチェックしているから保留でもいい。
  return {
    accessToken: body.access_token,//snake_case (access_token)からJS使われるるcamelCaseへ変換。
    refreshToken: body.refresh_token,
    email,
  };
}

export function refreshTokens(rt) {
  //refreshはユーザーフローでは使わないがsmokeテストで使う。
  const res = http.post(
    `${BASE_URL}/users/refresh`,
    JSON.stringify({ refresh_token: rt }),
    { headers: { "Content-Type": "application/json" }, tags: { name: "POST /users/refresh" } },
  );

  if (res.status !== 200) return null;

  const body = res.json();
  return {
    accessToken: body.access_token,
    refreshToken: body.refresh_token,
  };
}

export function params(token, name) {
  return {
    headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
    tags: { name },
  };
}//これは認証の責務やからユーザーシナリオに置かない。
//TODO:ただtagsはユーザーシナリオに置いたほうが良い。