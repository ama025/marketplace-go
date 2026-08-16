

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
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.TARGET_URL || 'http://localhost:9001';

export default function () {

  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health: status 200': (r) => r.status === 200,
  });

  const itemsRes = http.get(`${BASE_URL}/api/v1/items`);
  check(itemsRes, {
    'items: status 200': (r) => r.status === 200,
  });

  const brandsRes = http.get(`${BASE_URL}/api/v1/brands`);
  check(brandsRes, {
    'brands: status 200': (r) => r.status === 200,
  });

  const categoriesRes = http.get(`${BASE_URL}/api/v1/categories`);
  check(categoriesRes, {
    'categories: status 200': (r) => r.status === 200,
  });

  sleep(1);
}
