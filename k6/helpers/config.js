//今回はローカルからローカルとvpsへのk6負荷試験行うからenvを2つ用意する。
const rawUrl = __ENV.BASE_URL || "http://localhost:18000";
export const BASE_URL = rawUrl.replace(/\/$/, "");
// /が重複するとだるいからnormalize。
export const SEED_USERS = parseInt(__ENV.SEED_USERS || "500", 10);//第二引数で10進数を定義。
export const SEED_PREFIX = __ENV.SEED_PREFIX || "seed";

if (!__ENV.SEED_PASSWORD) {
  throw new Error("SEED_PASSWORD is required. Set it in k6/.env.k6.<env> or pass via --env");
}
export const SEED_PASSWORD = __ENV.SEED_PASSWORD;//これは存在しなければフォールバックではなく止める。

//これはパラメータではない。env使わない。
//これはloadテストで使う。k6実行で出力される。loadテストのthresholdは厳しく。
export const THRESHOLDS = {
  http_req_duration: ["p(95)<500", "p(99)<1000"],
  http_req_failed: ["rate<0.01"],
  checks: ["rate>0.99"],
};
//smokeテスト用。全チェック成功・エラーゼロ必須。
export const SMOKE_THRESHOLDS = {
  checks: ["rate==1.0"],
  http_req_failed: ["rate==0"],
};
//k6のconfigは責務違うから分ける。
