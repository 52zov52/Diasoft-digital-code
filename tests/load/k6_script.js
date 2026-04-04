import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const p95Latency = new Trend('p95_latency_ms', true);

export const options = {
  stages: [
    { duration: '30s', target: 2000 },  // Ramp-up
    { duration: '2m',  target: 10000 }, // Sustained 10k RPS
    { duration: '30s', target: 0 },     // Ramp-down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500'],      // SLA: 95% запросов < 500ms
    'http_req_failed': ['rate<0.01'],         // <1% ошибок
    'errors': ['rate<0.05'],
    'p95_latency_ms': [{ threshold: 'avg<450', abortOnFail: true }],
  },
  scenarios: {
    verify_manual: { executor: 'ramping-vus', gracefulStop: '5s', stages: options.stages, exec: 'verifyManual' },
    verify_qr:     { executor: 'ramping-vus', gracefulStop: '5s', stages: options.stages, exec: 'verifyQR', startTime: '5s' },
  }
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TOKENS = ['token-1', 'token-2', 'token-3', 'token-4', 'token-5'];

export function verifyManual() {
  const res = http.get(`${BASE_URL}/api/v1/verify?number=DIP${Math.floor(Math.random()*100000)}&university_code=UNI`);
  const passed = check(res, { 'status 200': (r) => r.status === 200 });
  errorRate.add(!passed);
  p95Latency.add(res.timings.duration);
  sleep(0.1);
}

export function verifyQR() {
  const token = TOKENS[Math.floor(Math.random() * TOKENS.length)];
  const res = http.get(`${BASE_URL}/api/v1/verify/qr/${token}`);
  const passed = check(res, { 'status 200': (r) => r.status === 200 });
  errorRate.add(!passed);
  p95Latency.add(res.timings.duration);
  sleep(0.1);
}