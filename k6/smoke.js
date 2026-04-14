import { check } from "k6";
import http from "k6/http";
import { BASE_URL, SMOKE_THRESHOLDS } from "./helpers/config.js";
import { login, refreshTokens, params } from "./helpers/auth.js";

//options は k6 のテスト全ての設定を定義するオブジェクト
export const options = {//smoketestはvuもiterationも1。
  vus: 1,
  iterations: 1,
  thresholds: SMOKE_THRESHOLDS,
};

export default function () {

  const health = http.get(`${BASE_URL}/health`, { tags: { name: "GET /health" } });
  check(health, { "health 200": (r) => r.status === 200 });
  //checkの構文。check(対象, { "名前": 判定関数, ... })

  const session = login(1);
  check(session, { "login succeeded": (s) => s !== null });
  if (!session) return;
  //session が falsyで終了。

  const refreshed = refreshTokens(session.refreshToken);
  check(refreshed, { "refresh succeeded": (r) => r !== null });
  if (!refreshed) return;

  //refreshとって終わりではなくrefreshtokenでlist叩けるかチェック。
  const token = refreshed.accessToken;

  //TODO:これ listWithRefreshed じゃなくてlistでいい。
const listWithRefreshed = http.get(
  `${BASE_URL}/tasks?limit=30`,
  params(token, "GET /tasks"),
);
check(listWithRefreshed, { "refreshed token works": (r) => r.status === 200 });

  const created = http.post(
    `${BASE_URL}/tasks`,
    JSON.stringify({ title: "k6 smoke test", description: "Smoke test task", due_date: 7 }),
    params(token, "POST /tasks"),
  );
  check(created, { "create 201": (r) => r.status === 201 });
  if (created.status !== 201) return;

  const taskId = created.json().id;

//versionはハードコードではなくresponseから受け取る。
  const afterCreate = http.get(`${BASE_URL}/tasks?limit=30`, params(token, "GET /tasks"));
  const createdTask = afterCreate.json().items.find((t) => t.id === taskId);
  if (!createdTask) return;

  const updated = http.patch(
    `${BASE_URL}/tasks/${taskId}`,
    JSON.stringify({ version: createdTask.version, title: "k6 smoke updated" }),
    params(token, "PATCH /tasks/{id}"),
  );
  check(updated, { "update 200": (r) => r.status === 200 });
  if (updated.status !== 200) return;


  const afterUpdate = http.get(`${BASE_URL}/tasks?limit=30`, params(token, "GET /tasks"));
  const updatedTask = afterUpdate.json().items.find((t) => t.id === taskId);
  if (!updatedTask) return;

  const deleted = http.del(
    `${BASE_URL}/tasks/${taskId}`,
    JSON.stringify({ version: updatedTask.version }),
    params(token, "DELETE /tasks/{id}"),
  );
  check(deleted, { "delete 204": (r) => r.status === 204 });
}//一連のsmokeでtaskが消える->冪等。
