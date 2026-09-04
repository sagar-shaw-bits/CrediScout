// CrediScout load test (k6).
//
// Two scenarios, selected via the SCENARIO env var:
//   healthz : cheap path — measures the server's raw request-handling
//             headroom, independent of the LLM. Does NOT need Ollama.
//   ask     : full RAG path (embed -> hybrid retrieve -> LLM). Latency is
//             dominated by Ollama generation. Needs Ollama + a built index.
//
// Run:
//   k6 run scripts/load-test.js                      # default: healthz
//   k6 run -e SCENARIO=healthz scripts/load-test.js
//   k6 run -e SCENARIO=ask scripts/load-test.js
//
// Tunables (env vars):
//   BASE_URL   default http://localhost:8080
//   VUS        concurrent virtual users            (default 50 healthz / 10 ask)
//   DURATION   how long to sustain load            (default 30s / 1m)
//   QUESTION   question for the ask scenario
//   TOP_K      chunks retrieved per question       (default 5)

import http from 'k6/http';
import { check } from 'k6';
import { Trend } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const SCENARIO = (__ENV.SCENARIO || 'healthz').toLowerCase();
const QUESTION = __ENV.QUESTION || 'What are the benefits of the Regalia card?';
const TOP_K = parseInt(__ENV.TOP_K || '5', 10);

const askLatency = new Trend('ask_latency_ms', true);

function opts() {
  if (SCENARIO === 'ask') {
    return {
      scenarios: {
        ask: {
          executor: 'constant-vus',
          vus: parseInt(__ENV.VUS || '10', 10),
          duration: __ENV.DURATION || '1m',
        },
      },
      thresholds: {
        http_req_failed: ['rate<0.01'],   // <1% errors
        http_req_duration: ['p(95)<60000'], // end-to-end incl. LLM
      },
    };
  }
  // healthz
  return {
    scenarios: {
      healthz: {
        executor: 'constant-vus',
        vus: parseInt(__ENV.VUS || '50', 10),
        duration: __ENV.DURATION || '30s',
      },
    },
    thresholds: {
      http_req_failed: ['rate<0.001'],
      http_req_duration: ['p(95)<50', 'p(99)<150'],
    },
  };
}

export const options = opts();

export default function () {
  if (SCENARIO === 'ask') {
    const payload = JSON.stringify({ question: QUESTION, top_k: TOP_K });
    const res = http.post(`${BASE_URL}/api/ask`, payload, {
      headers: { 'Content-Type': 'application/json' },
      timeout: '120s',
    });
    askLatency.add(res.timings.duration);
    check(res, {
      'status 200': (r) => r.status === 200,
      'has answer': (r) => {
        try {
          return typeof JSON.parse(r.body).answer === 'string';
        } catch (_) {
          return false;
        }
      },
    });
  } else {
    const res = http.get(`${BASE_URL}/healthz`);
    check(res, {
      'status 200': (r) => r.status === 200,
      'status ok': (r) => {
        try {
          return JSON.parse(r.body).status === 'ok';
        } catch (_) {
          return false;
        }
      },
    });
  }
}
