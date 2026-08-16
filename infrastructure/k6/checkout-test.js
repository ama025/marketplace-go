

import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 5 },
    { duration: '1m',  target: 5 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'],
    http_req_failed: ['rate<0.05'],
  },
};

const BASE_URL = __ENV.TARGET_URL || 'http://localhost:9004';

export default function () {

  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health: status 200': (r) => r.status === 200,
  });

  const createRes = http.post(
    `${BASE_URL}/api/v1/orders`,
    JSON.stringify({
      AccountName: `k6-user-${__VU}`,
      Items: [
        {
          ItemID: '00000000-0000-0000-0000-000000000001',
          ItemTitle: 'k6 Test Product',
          Quantity: 1,
          UnitPrice: 49.99,
          Discount: 10,
        },
      ],
    }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  check(createRes, {
    'create order: 201': (r) => r.status === 201,
  });

  const listRes = http.get(`${BASE_URL}/api/v1/orders?account=k6-user-${__VU}`);
  check(listRes, {
    'list orders: 200': (r) => r.status === 200,
  });

  sleep(1);
}
