import { userFlow } from "./helpers/flow.js";

export const options = {
  scenarios: {
    stress: {
      executor: "ramping-arrival-rate",
      startRate: 10,
      timeUnit: "1s",
      stages: [//10 → 1000。8 分 (2m × 4 ステージ)
        { duration: "2m", target: 200 },
        { duration: "2m", target: 500 },
        { duration: "2m", target: 1000 },
        { duration: "2m", target: 0 },//壊れた後の戻りを確かめる。
      ],
      preAllocatedVUs: 200,
      maxVUs: 2000,
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<2000"],
    http_req_failed: ["rate<0.05"],
  },
};

export default function () {
  userFlow();
}
