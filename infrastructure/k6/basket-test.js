

import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 10 },
    { duration: '1m',  target: 10 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.05'],
  },
};

const BASE_URL = __ENV.TARGET_URL || 'http://localhost:9002';

export default function () {

  const account = `k6-user-${__VU}-${__ITER}`;

  const saveRes = http.post(
    `${BASE_URL}/api/v1/cart`,
    JSON.stringify({
      accountName: account,
      items: [
        {
          itemId: '00000000-0000-0000-0000-000000000001',
          quantity: 2,
          unitPrice: 99.90,
          itemTitle: 'k6 Test Item',
        },
      ],
    }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  check(saveRes, {
    'save cart: 204': (r) => r.status === 204,
  });

  const getRes = http.get(`${BASE_URL}/api/v1/cart/${account}`);
  check(getRes, {
    'get cart: 200': (r) => r.status === 200,
    'get cart: has items': (r) => {
      const body = JSON.parse(r.body);
      return body.items && body.items.length > 0;
    },
  });

  const delRes = http.del(`${BASE_URL}/api/v1/cart/${account}`);
  check(delRes, {
    'delete cart: 204': (r) => r.status === 204,
  });

  sleep(1);
}
