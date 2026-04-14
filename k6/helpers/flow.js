import { check } from "k6";
import http from "k6/http";
import exec from "k6/execution";
import { login } from "./auth.js";
import { params } from "./auth.js";
import { BASE_URL, SEED_USERS } from "./config.js";

//tokenはVUの中で共有。arrival-rateでは別VUは別ランタイム。
//これがなければiterationごとにloginが動く。これは実際の挙動とは違う。
let token = null;

function seedUserNum() {
  return ((exec.vu.idInTest - 1) % SEED_USERS) + 1;
}//exec.vu.idInTestは一意。これ多く使われる記法。%で循環。


export function ensureLogin() {
  if (!token) {
    const session = login(seedUserNum());
    if (!check(session, { "login succeeded": (s) => s !== null })) {
      exec.test.abort(`login failed for user ${seedUserNum()}`);
    }//失敗はオプショナルチェイニングでundefinedではなくcheck+abort
    token = session.accessToken;
  }
  return token;
}

export function userFlow() {
  const t = ensureLogin();

  //累積確率使う。30個のタスクでdelete続いて枯渇することはないと考えている。
  //TODO:delete updateの前にtask存在するかチェック。
  const rand = Math.random();
  if (rand < 0.60) {
    listTasks(t);
  } else if (rand < 0.80) {
    createTask(t);
  } else if (rand < 0.95) {
    updateRandomTask(t);
  } else {
    deleteRandomTask(t);
  }
}

function listTasks(t) {
  const res = http.get(`${BASE_URL}/tasks?limit=30`, params(t, "GET /tasks"));//limit5は弱い。実際は30ぐらい。
  if (res.status === 401) { token = null; return; }
  check(res, { "list 200": (r) => r.status === 200 });
}

function createTask(t) {
  const body = JSON.stringify({
    title: "k6 load test task",
    description: "Created during load testing",
    due_date: 7,
  });
  const res = http.post(`${BASE_URL}/tasks`, body, params(t, "POST /tasks"));
  if (res.status === 401) { token = null; return; }
  check(res, { "create 201": (r) => r.status === 201 });
}

function updateRandomTask(t) {
  const list = http.get(`${BASE_URL}/tasks?limit=30`, params(t, "GET /tasks"));
  if (list.status !== 200) { if (list.status === 401) token = null; return; }
  const items = list.json().items;
  if (!items || items.length === 0) return;
  const task = items[Math.floor(Math.random() * items.length)];

  const body = JSON.stringify({ version: task.version, title: "k6 updated" });
  const res = http.patch(`${BASE_URL}/tasks/${task.id}`, body, params(t, "PATCH /tasks/{id}"));
  if (res.status === 401) { token = null; return; }
  check(res, { "update 200|409": (r) => r.status === 200 || r.status === 409 });
}

function deleteRandomTask(t) {
  const list = http.get(`${BASE_URL}/tasks?limit=30`, params(t, "GET /tasks"));
  if (list.status !== 200) { if (list.status === 401) token = null; return; }
  const items = list.json().items;
  if (!items || items.length === 0) return;
  const task = items[Math.floor(Math.random() * items.length)];

  const res = http.del(
    `${BASE_URL}/tasks/${task.id}`,
    JSON.stringify({ version: task.version }),
    params(t, "DELETE /tasks/{id}"),
  );
  if (res.status === 401) { token = null; return; }
  check(res, { "delete 204|409": (r) => r.status === 204 || r.status === 409 });
}
