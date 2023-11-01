import { check, fail } from 'k6';
import loki from 'k6/x/loki';

/**
 * URL used for push and query requests
 * Path is automatically appended by the client
 * Example: https://tenantID@localhost:3100
 * @constant {string}
 */
const LOKI_URL = __ENV.LOKI_URL || fail("provide LOKI_URL when starting k6");


/**
 * K6 timeout when call Loki.
 * This should be less than http_server_write_timeout https://github.com/grafana/loki/blob/v2.9.1/docs/sources/configure/_index.md?plain=1#L315
 * @type {number}
 */
const timeout = parseInt('30000'); // 30000ms = 30s

/**
 * The ratio between JSON and Protobuf encoded batch requests for pushing logs.
 * @type {number}
 */
const ratio = parseFloat('0.9');

/**
 * Total number of cardinality is 50 * 200 * 10000 * 400 * 7 * 2 * 200 = 1.12E14
 * This is high cardinality. This can kill Loki.
 *
 * https://grafana.com/blog/2020/04/21/how-labels-in-loki-can-make-log-queries-faster-and-easier/
 *
 * @type {{namespace: number, app: number, pod: number, container: number, format: number, stream: number, instance: number}}
 */
const labelCardinality = {
    "namespace": 50,
    "app": 200,
    "pod": 10000,
    "container": 400,
    "format": 7,
    "stream": 2,
    "instance": 200,
}

// overriding the instance label to ensure we have control over cardinality
// https://github.com/grafana/xk6-loki/blob/12ba135193ecb17f37d043262f2f145d5b9cf641/batch.go#L170
// https://github.com/grafana/xk6-loki/blob/12ba135193ecb17f37d043262f2f145d5b9cf641/examples/custom-labels.js#L19
let labels = loki.Labels({
    "format": ["apache_common","apache_combined","apache_error","rfc3164","rfc5424","json","logfmt"],
    "stream": ["stdout","stderr"],
    "instance": ["ap-southeast-225.94.154.207","ap-southeast-19.234.144.150","ap-southeast-113.174.208.146","ap-southeast-240.56.62.40","ap-southeast-142.7.124.245","ap-southeast-75.81.155.80","ap-southeast-99.141.124.42","ap-southeast-51.79.13.59","ap-southeast-39.78.178.193","ap-southeast-98.4.111.69","ap-southeast-80.254.110.237","ap-southeast-82.246.20.33","ap-southeast-70.215.255.29","ap-southeast-203.193.21.145","ap-southeast-71.216.128.27","ap-southeast-140.13.76.17","ap-southeast-245.127.218.249","ap-southeast-142.83.59.52","ap-southeast-32.148.208.196","ap-southeast-20.219.248.66","ap-southeast-82.24.180.205","ap-southeast-80.21.187.102","ap-southeast-193.178.97.123","ap-southeast-134.74.128.151","ap-southeast-39.254.235.248","ap-southeast-145.140.117.150","ap-southeast-161.145.37.166","ap-southeast-16.83.47.168","ap-southeast-152.100.45.55","ap-southeast-233.241.99.187","ap-southeast-225.252.181.15","ap-southeast-217.85.73.16","ap-southeast-243.197.199.239","ap-southeast-106.151.88.127","ap-southeast-114.146.164.20","ap-southeast-3.226.161.129","ap-southeast-221.159.179.18","ap-southeast-205.36.209.116","ap-southeast-84.32.202.245","ap-southeast-27.121.249.32","ap-southeast-165.92.138.36","ap-southeast-149.21.11.90","ap-southeast-252.98.84.244","ap-southeast-197.64.144.150","ap-southeast-94.68.41.50","ap-southeast-39.157.241.96","ap-southeast-58.182.118.142","ap-southeast-70.199.155.55","ap-southeast-176.159.161.253","ap-southeast-63.121.73.158","ap-southeast-193.155.101.155","ap-southeast-205.88.28.243","ap-southeast-104.229.184.89","ap-southeast-52.41.38.231","ap-southeast-208.83.238.231","ap-southeast-176.39.168.31","ap-southeast-180.237.21.13","ap-southeast-37.35.155.1","ap-southeast-135.133.109.83","ap-southeast-108.235.152.128","ap-southeast-17.83.245.137","ap-southeast-180.158.133.253","ap-southeast-70.177.216.103","ap-southeast-177.9.160.161","ap-southeast-195.59.138.189","ap-southeast-23.33.209.251","ap-southeast-205.50.129.114","ap-southeast-45.0.169.62","ap-southeast-28.3.216.5","ap-southeast-4.90.57.165","ap-southeast-66.82.191.197","ap-southeast-52.32.86.32","ap-southeast-210.57.35.124","ap-southeast-140.85.119.98","ap-southeast-24.223.148.95","ap-southeast-55.71.30.125","ap-southeast-180.61.28.50","ap-southeast-141.202.230.21","ap-southeast-213.40.153.20","ap-southeast-105.232.109.48","ap-southeast-154.215.85.198","ap-southeast-124.50.114.229","ap-southeast-67.75.107.241","ap-southeast-144.155.71.109","ap-southeast-30.189.52.196","ap-southeast-19.139.11.97","ap-southeast-251.65.131.125","ap-southeast-77.182.221.180","ap-southeast-130.101.23.33","ap-southeast-36.84.121.36","ap-southeast-225.149.31.43","ap-southeast-193.235.139.189","ap-southeast-86.13.184.56","ap-southeast-54.177.165.152","ap-southeast-53.76.243.229","ap-southeast-70.229.216.86","ap-southeast-18.39.103.189","ap-southeast-244.4.196.134","ap-southeast-193.179.87.31","ap-southeast-110.128.88.139","ap-southeast-54.234.208.188","ap-southeast-235.251.94.3","ap-southeast-25.114.209.37","ap-southeast-33.29.154.207","ap-southeast-67.205.218.186","ap-southeast-56.70.237.174","ap-southeast-233.73.182.23","ap-southeast-240.1.220.219","ap-southeast-77.94.37.197","ap-southeast-152.189.72.97","ap-southeast-83.30.12.190","ap-southeast-218.255.155.19","ap-southeast-116.100.122.179","ap-southeast-111.217.135.234","ap-southeast-146.230.49.84","ap-southeast-106.188.223.137","ap-southeast-10.253.154.139","ap-southeast-200.158.66.190","ap-southeast-162.229.178.0","ap-southeast-36.90.162.42","ap-southeast-7.175.100.146","ap-southeast-5.185.104.155","ap-southeast-219.131.90.50","ap-southeast-114.115.237.125","ap-southeast-38.216.243.53","ap-southeast-184.28.26.127","ap-southeast-116.251.118.36","ap-southeast-136.101.226.234","ap-southeast-201.236.112.21","ap-southeast-5.122.175.86","ap-southeast-3.23.22.97","ap-southeast-58.156.143.25","ap-southeast-121.225.131.126","ap-southeast-60.222.162.252","ap-southeast-216.30.191.9","ap-southeast-123.187.136.193","ap-southeast-250.54.173.140","ap-southeast-91.135.168.10","ap-southeast-245.95.227.182","ap-southeast-72.109.222.73","ap-southeast-2.213.22.32","ap-southeast-179.169.138.113","ap-southeast-206.71.160.23","ap-southeast-58.109.178.8","ap-southeast-155.144.114.154","ap-southeast-9.158.148.18","ap-southeast-188.53.68.221","ap-southeast-56.159.33.193","ap-southeast-112.243.180.26","ap-southeast-167.190.254.92","ap-southeast-43.81.53.127","ap-southeast-70.66.27.188","ap-southeast-198.52.131.46","ap-southeast-71.109.218.24","ap-southeast-193.248.27.105","ap-southeast-151.74.113.207","ap-southeast-162.247.50.61","ap-southeast-182.104.97.70","ap-southeast-31.156.167.49","ap-southeast-87.81.246.42","ap-southeast-62.224.123.226","ap-southeast-175.117.150.195","ap-southeast-111.202.237.44","ap-southeast-102.188.153.35","ap-southeast-185.60.171.36","ap-southeast-132.26.67.166","ap-southeast-216.90.144.161","ap-southeast-46.172.73.107","ap-southeast-114.99.46.135","ap-southeast-253.159.221.75","ap-southeast-91.231.138.122","ap-southeast-195.194.46.4","ap-southeast-157.152.157.137","ap-southeast-120.186.193.110","ap-southeast-48.29.95.234","ap-southeast-3.50.176.102","ap-southeast-145.206.82.61","ap-southeast-238.211.126.250","ap-southeast-70.24.17.197","ap-southeast-248.8.226.172","ap-southeast-220.18.96.228","ap-southeast-46.255.54.97","ap-southeast-157.138.0.61","ap-southeast-247.120.200.196","ap-southeast-150.139.211.235","ap-southeast-29.214.106.180","ap-southeast-205.29.11.25","ap-southeast-18.225.176.62","ap-southeast-172.253.194.201","ap-southeast-106.15.245.193","ap-southeast-138.210.245.144","ap-southeast-57.145.252.164","ap-southeast-111.32.2.4","ap-southeast-170.57.141.126","ap-southeast-220.22.151.93","ap-southeast-66.222.62.161","ap-southeast-200.160.29.160","ap-southeast-75.137.189.211","ap-southeast-79.119.1.246","ap-southeast-7.151.39.16"],
});

let namespace = [];
for (let i = 1; i <= labelCardinality.namespace; i++) {
    namespace[i-1] = `ns-${zeroPad(i, 2)}`;
}
labels.namespace = namespace;

let app = [];
for (let i = 1; i <= labelCardinality.app; i++) {
    app[i-1] = `service-name-resemble-helm-chart-name-${zeroPad(i, 3)}`;
}
labels.app = app;

let pod = [];
for (let i = 1; i <= labelCardinality.pod; i++) {
    pod[i-1] = `service-name-pod-name-${zeroPad(i, 5)}`;
}
labels.pod = pod;

let container = [];
let containerID = 1;
for (let i = 1; i <= labelCardinality.container; i++) {
    if (i % 2 === 0) {
        container[i-1] = `sidecar-${zeroPad(containerID, 3)}`;
        containerID++
    } else {
        container[i-1] = `application-${zeroPad(containerID, 3)}`;
    }
}
labels.container = container;

function zeroPad(num, places) {
    let zero = places - num.toString().length + 1;
    return Array(+(zero > 0 && zero)).join("0") + num;
}

/**
 * setup Set up data for processing, share data among VUs.
 * Example: Call API to start test environment.
 * This will be called once when k6 running.
 *
 * Read about k6 lifecycle here https://github.com/grafana/k6-docs/blob/v0.47.5/src/data/markdown/translated-guides/en/02%20Using%20k6/06%20Test%20lifecycle.md
 */
export function setup() {
    console.log(`Write test started at: ${new Date().toISOString()}`)
}

/**
 * teardown Process result of setup code, stop test environment
 * Example: Validate that setup had a certain result, send webhook notifying that test has finished.
 * This will be called once when k6 finished.
 * @param data
 */
export function teardown(data) {
    console.log(`Write test completed at: ${new Date().toISOString()} with data ${data}`)
}

/**
 * Constants for byte values
 * @constant {number}
 */
const KB = 1024;

/**
 * Definition of test scenario
 */
export const options = {
    // https://github.com/grafana/k6-docs/blob/v0.47.0/src/data/markdown/translated-guides/en/02%20Using%20k6/04%20Thresholds.md
    thresholds: {
        'http_req_failed': [
            { threshold: 'rate<=0.01' }, // http errors should be less than 1%
        ],
        'http_req_duration': [
            { threshold: 'p(95)<500' }, // 95% of requests should be below 500ms
        ],
    },
    scenarios: {
        constant_write: {
            executor: 'constant-arrival-rate',
            exec: 'write',
            rate: 100,
            timeUnit: '1s', // 100 request (rate) per second
            duration: '2m',
            preAllocatedVUs: 500,
            maxVUs: 500,
        },
    },
};

const conf = new loki.Config(`${LOKI_URL}`.trim(), timeout, ratio, labelCardinality, labels);
const client = new loki.Client(conf);

/**
 * Entrypoint for write scenario.
 * This will push log about 5 Mbps up to 7 Mbps with 8 - 10 streams per batch.
 * minSize kb * request per second = bytes per second = 50 KB * 100 rps = 5000 KB/s = 5 MB/s
 * maxSize kb * request per second = bytes per second = 70 KB * 100 rps = 7000 KB/s = 7 MB/s
 *
 * In Promtail, maximum default batch size is 1MB and batch wait is 1s.
 * So, even if we don't have 1MB full in 1s, it will be pushed by Promtail to Loki, and if we have more than 1MB per second,
 * it will be split to 1MB batch.
 * https://github.com/grafana/loki/blob/v2.9.2/docs/sources/send-data/promtail/configuration.md?plain=1#L236-L242
 *
 * In other words, we don't need to increase this number for load testing.
 */
export function write() {
    let streams = randomInt(8, 10); // The amount of streams the pushed batch should contain.
    let minSize = parseInt((50 * KB).toString());
    let maxSize = parseInt((70 * KB).toString());
    let res = client.pushParameterized(streams, minSize, maxSize);
    check(res,
        {
            'successful write': (res) => {
                let success = res.status === 204;
                if (!success) {
                    console.log('Unsuccessful', res.status, res.body)
                }
                return success;
            },
        }
    );
}

/**
 * Return a random integer between min and max including min and max
 */
function randomInt(min, max) {
    return Math.floor(Math.random() * (max - min + 1) + min);
}
