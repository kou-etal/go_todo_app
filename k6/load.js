import { THRESHOLDS } from "./helpers/config.js";
import { userFlow } from "./helpers/flow.js";

export const options = {
  scenarios: {
    //複数シナリオを並列で動かすこともできる。arrival-rate使う。
    load: {
      executor: "constant-arrival-rate",
      rate: 300,//これはrpsではなくiterations per second
      timeUnit: "1s",
      duration: "5m",
      preAllocatedVUs: 100,
      maxVUs: 500,
    },
  },
  thresholds: THRESHOLDS,
};
//k6 はdefault exportを認識する。
export default function () {
  userFlow();
}
