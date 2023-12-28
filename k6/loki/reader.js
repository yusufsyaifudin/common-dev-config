import { URL } from 'https://jslib.k6.io/url/1.0.0/index.js';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import { check, fail, sleep } from 'k6';
import http from 'k6/http';

/**
 * URL used for push and query requests
 * Path is automatically appended by the client
 * Example: https://tenantID@localhost:3100
 * @constant {string}
 */
const LOKI_URL = __ENV.LOKI_URL || fail("provide LOKI_URL when starting k6");

/**
 * Tenant ID to query data.
 */
const TENANT_ID = __ENV.TENANT_ID || fail("provide TENANT_ID when starting k6")

/**
 * Limit of returned rows, should under volume_max_series configuration.
 * @type {number}
 */
const LIMIT = 1000;

/**
 * List of application that exist under the same tenant ID.
 * Get using this API:
 * curl -X GET 'http://127.0.0.1:3100/loki/api/v1/label/app/values' -H 'X-Scope-OrgID: tenant-id'
 * @type {string[]}
 */
let apps = [
    "argo-rollouts",
    "argocd-application-controller",
    "argocd-applicationset-controller",
    "argocd-notifications-controller",
    "argocd-repo-server",
    "argocd-server",
    "writer-log-2hour",
];

/**
 * setup Set up data for processing, share data among VUs.
 * Example: Call API to start test environment.
 * This will be called once when k6 running.
 *
 * Read about k6 lifecycle here https://github.com/grafana/k6-docs/blob/v0.47.5/src/data/markdown/translated-guides/en/02%20Using%20k6/06%20Test%20lifecycle.md
 */
export function setup() {
   console.log(`Read test started at: ${new Date().toISOString()}`)
}

/**
 * teardown Process result of setup code, stop test environment
 * Example: Validate that setup had a certain result, send webhook notifying that test has finished.
 * This will be called once when k6 finished.
 * @param data
 */
export function teardown(data) {
    console.log(`Read test completed at: ${new Date().toISOString()} with data ${data}`)
}

/**
 * Definition of test scenario.
 * If you encounter error such as "Insufficient VUs, reached 100 active VUs and cannot initialize more"
 * Try read this: https://stackoverflow.com/a/76329768
 */
export const options = {
    // https://github.com/grafana/k6-docs/blob/v0.47.0/src/data/markdown/translated-guides/en/02%20Using%20k6/04%20Thresholds.md
    thresholds: {
        'http_req_failed': [
            { threshold: 'rate<=0.01' }, // http errors should be less than 1%
        ],
        'http_req_duration': [
            { threshold: 'p(95)<60000' }, // 95% of requests should be below 60s
        ],
    },
    scenarios: {
        constant_read: {
            executor: 'constant-vus',
            exec: 'readRange',
            vus: 8,
            duration: '30m',
        },
    },
};


const rangesRatioConfig = [
    {
        ratio: 0.1,
        item: '15m'
    },
    {
        ratio: 0.1,
        item: '30m'
    },
    {
        ratio: 0.1,
        item: '1h'
    },
    {
        ratio: 0.1,
        item: '3h'
    },
    {
        ratio: 0.1,
        item: '12h'
    },
    {
        ratio: 0.1,
        item: '24h'
    },
    {
        ratio: 0.1,
        item: '48h' // 2 days
    },
    {
        ratio: 0.1,
        item: '72h' // 3 days
    },
    {
        ratio: 0.1,
        item: '168h' // 7 days
    },
    {
        ratio: 0.1,
        item: '720h' // 30 days
    },
];

/**
 * createSelectorByRatio will return time range, such as 15m, 1h, etc
 * @param ratioConfig
 * @returns {(function(*): (*|string))|*}
 */
const createSelectorByRatio = (ratioConfig) => {
    let ratioSum = 0;
    const executorsIntervals = [];
    for (let i = 0; i < ratioConfig.length; i++) {
        executorsIntervals.push({
            start: ratioSum,
            end: ratioSum + ratioConfig[i].ratio,
            item: ratioConfig[i].item,
        })
        ratioSum += ratioConfig[i].ratio
    }
    return (random) => {
        if (random >= 1 || random < 0) {
            fail(`random value must be within range [0-1)`)
        }
        const value = random * ratioSum;
        for (let i = 0; i < executorsIntervals.length; i++) {
            let currentInterval = executorsIntervals[i];
            if (value < currentInterval.end && value >= currentInterval.start) {
                return currentInterval.item
            }
        }
    }
}

const selectRangeByRatio = createSelectorByRatio(rangesRatioConfig);

const rangeQuerySuppliers = [
    // Type 1: without filter
    () => `{app="${randomChoice(apps)}"}`, // query all with app name label

    // Type 2: substring
    () => `{app="${randomChoice(apps)}"} |= "a"`, // search line that contains "a", why "a"? because most log contains "a" and we want to get the result

    // Type 3: Regex
    () => `{app="${randomChoice(apps)}"} |~ "HTTP/.*(1|2)"`, // search regex with HTTP/1{anything} or HTTP/2{anything}

    // Type 4: JSON parser
    () => `{app="${randomChoice(apps)}"} | json level="level" | level = "ERROR"`, // query with json parser and level ERROR
    () => `{app="${randomChoice(apps)}"} | json level="level" | level = "error"`,

    // Type 5: logfmt parser
    () => `{app="${randomChoice(apps)}"} | logfmt | level = "ERROR"`,
    () => `{app="${randomChoice(apps)}"} | logfmt | level = "error"`, // example on {app="argocd-server"} | logfmt | level = `error`
]

/**
 * Execute range query with given client.
 * Equivalent with:
 * curl -X GET 'http://127.0.0.1:3100/loki/api/v1/query_range?query={app%3D%22writer-log-2hour%22}%20%7C%3D%20%22ERROR%22&direction=backward&limit=1000&since=30d' \
 *   -H 'X-Scope-OrgID: tenant-id'
 */
export function readRange() {
    // Randomly select the query supplier from the pool
    // and call the supplier that provides prepared query.
    const query = randomChoice(rangeQuerySuppliers)();
    // Randomly select the range.
    let range = selectRangeByRatio(Math.random());

    const url = new URL(`${LOKI_URL}/loki/api/v1/query_range`);

    url.searchParams.append('direction', 'backward');
    url.searchParams.append('limit', `${LIMIT}`);
    url.searchParams.append('query', query);
    url.searchParams.append('since', range);

    const res = http.get(url.toString(), {
        headers: {
            'X-Scope-OrgID': TENANT_ID,
        },
        timeout: '60s',
        redirects: 3,
    });

    // Assert the response from loki.
    checkResponse(res, "successful range query", range);

    // Sleep for 500 ms to 1000 ms (1s) before next request.
    // Like in real use-case, the developers may not click the query button many times once it's done.
    // They expected to read the logs result, analyze it until they realize that it is not met their expectation and then do another query.
    let delay = randomIntBetween(5, 10) / 10;
    sleep(delay);
}


const checkResponse = (response, name, range) => {
    const checkName = `${name}[${range}]`;
    const assertion = {};
    assertion[range ? checkName : name] = (res) => {
        let success = res.status === 200;
        if (!success) {
            console.log('Unsuccessful query', res.status, res.body);
        }
        return success;
    };
    check(response, assertion);
}

/**
 * Return an item of random choice of a list
 */
function randomChoice(items) {
    return items[Math.floor(Math.random() * items.length)];
}