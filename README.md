imUp client
--

![CI](https://github.com/imup-io/client/workflows/CI/badge.svg)
![S](https://github.com/imup-io/client/workflows/CodeQL/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/imup-io/client.svg)](https://pkg.go.dev/github.com/imup-io/client)

## imUp

This application is the client binary distributed for use with [imUp.io's](https://imUp.io) connectivity and speed testing platform

It continuously checks for connectivity and periodically runs speed tests in accordance with the [ndt7-protocol](https://github.com/m-lab/ndt-server/blob/master/spec/ndt7-protocol.md#requirements-for-non-interactive-clients)

The client is also enabled with imUp real-time features including on-demand speed tests, liveness check-ins as well as remote configuration

## Features

- Connection Health: Collects latency, RTT, and packet loss. If you're experiencing connectivity issues while working from home or gaming, this data will help you pinpoint if the issue is on your end or not.

- Speed Tests: Using the ndt7 protocol collects speed tests data at semi random intervals or on-demand.

- Real-time Monitoring: With [imUp.io](https://imUp.io) you have access to on-demand speed tests as well as real-time altering via email or twitter.

## Usage

Download the client

   ```sh
   go install github.com/imup-io/client@latest
   ```

Run the client specifying an email address you want to associate the data with

  ```sh
  client --email email@example.com
  ```

## Environment Configuration

|        Name                        |      Description                          |                   Default                                    |
|------------------------------------|-------------------------------------------|--------------------------------------------------------------|
| `ALLOWLISTED_IPS`                  | configures allowed ips for speed tests    | `""`                                                         |
| `API_KEY`                          | api key for imup orgs                     | `""`                                                         |
| `BLOCKLISTED_IPS`                  | configures blocked ips for speed tests    | `""`                                                         |
| `CONN_DELAY`                       | time between dials in milliseconds        | `"200"`                                                      |
| `CONN_INTERVAL`                    | dialer interval in seconds                | `"60"`                                                       |
| `CONN_REQUESTS`                    | number of requests each test              | `"300"`                                                      |
| `EMAIL`                            | email address associated with imup data   | `""`                                                         |
| `GROUP_ID`                         | id associated with an imup org group      | `""`                                                         |
| `LOG_TO_FILE`                      | log output to a file in the default dir   | `"false"`                                                    |
| `IMUP_ADDRESS`                     | imup API address                          | `"https://api.imup.io/v1/data/connectivity"`                 |
| `IMUP_ADDRESS_SPEEDTEST`           | imup API address for speedtest            | `"https://api.imup.io/v1/data/speedtest"`                    |
| `IMUP_LIVENESS_CHECKIN_ADDRESS`    | imup API address for liveness checkin     | `"https://api.imup.io/v1/realtime/livenesscheckin"`          |
| `IMUP_SHOULD_RUN_SPEEDTEST_ADDRESS`| imup API address for on-demand speedtests | `"https://api.imup.io/v1/realtime/shouldClientRunSpeedTest"` |
| `IMUP_SPEED_TEST_RESULTS_ADDRESS`  | imup API address for speed test results   | `"https://api.imup.io/v1/realtime/speedTestResults"`         |
| `IMUP_SPEED_TEST_STATUS_ADDRESS`   | imup API address for speed tests running  | `"https://api.imup.io/v1/realtime/speedTestStatusUpdate"`    |
| `IMUP_REALTIME_AUTHORIZED`         | imup API address for real-time authorized | `"https://api.imup.io/v1/auth/real-timeAuthorized"`          |
| `IMUP_REALTIME_CONFIG`             | imup API address for reloadable config    | `"https://api.imup.io/v1/realtime/config"`                   |
| `IMUP_DATA_LENGTH`                 | imup data length per interval             | `"15"`                                                       |
| `PING_ADDRESS`                     | address to ping                           | `"1.1.1.1,1.0.0.1,8.8.8.8,8.8.4.4"` (CloudFlare /Google DNS) |
| `PING_ADDRESS_INTERNAL`            | configurable gateway address              | discovered/configurable (disabled with --no-discover-gateway)|
| `PING_DELAY`                       | time between pings in milliseconds        | `"100"`                                                      |
| `PING_INTERVAL`                    | ping interval in seconds                  | `"60"`                                                       |
| `PING_REQUESTS`                    | number of requests each test              | `"600"`                                                      |
| `SPEEDTEST_ENABLED`                | enable speed tests                        | `"false"`                                                    |
| `SPEEDTEST_INTERVAL`               | intended to set cron frequency            | VARIABLE CURRENTLY UNUSED                                    |
| `VERBOSITY`                        | controls log output                       | `"info"`                                                     |

## Flags

```sh
Usage: imupClient
  -allowlisted-ips string
     Allowed IPs for speed tests
  -anonymize.ip value
     Valid values are "none" and "netblock". (default none)
  -api-post-connection-data string
     default api endpoint is https://api.imup.io/v1/data/connectivity
  -api-post-speed-test-data string
     default api endpoint is https://api.imup.io/v1/data/speedtest
  -blocklisted-ips string
     Blocked IPs for speed tests
  -config-version string
     config version
  -conn-delay string
     default is to wait 200ms between each net conn
  -conn-interval string
     default is to run a conn test once every 60s
  -conn-requests string
     default is to send 300 requests each test
  -email string
     email address
  -group-id string
     org users group id
  -host-id string
     host id
  -imup-data-length string
     default is to collect 15 data points and then send data to the api
  -insecure
     run insecure speed tests (ws:// and not wss://)
  -key string
     api key
  -liveness-check-in-address string
     default api endpoint is https://api.imup.io/v1/realtime/livenesscheckin
  -locate.url value
     The base url for the Locate API (default https://locate.measurementlab.net/v2/nearest/)
  -log-to-file
     if enabled, will log to the default root directory to use for user-specific cached data
  -no-gateway-discovery
     do not attempt to discover a default gateway
  -no-speed-test
     don't run speed tests
  -nonvolatile
     use disk to store collected data between tests to ensure reliability
  -ping
     use ICMP ping for connectivity tests
  -ping-address-internal string
     client by default attempts to discover the internal gateway
  -ping-addresses-external string
     default addrs to test against are 1.1.1.1,1.0.0.1,8.8.8.8,8.8.4.4
  -ping-delay string
     default is to wait 100ms between each ping
  -ping-interval string
     default is to run a ping test once every 60s
  -ping-requests string
     default is to send 600 requests each test
  -realtime
     enable realtime features, enabled by default (default true)
  -realtime-authorized string
     default api endpoint is https://api.imup.io/v1/auth/realtimeAuthorized
  -realtime-config string
     default api endpoint is https://api.imup.io/v1/realtime/config
  -should-run-speed-test-address string
     default api endpoint is https://api.imup.io/v1/realtime/shouldClientRunSpeedTest
  -speed-test-results-address string
     default api endpoint is https://api.imup.io/v1/realtime/speedTestResults
  -speed-test-status-update-address string
     default api endpoint is https://api.imup.io/v1/realtime/speedTestStatusUpdate
  -verbosity string
     How verbose log output should be (Default Info)
```

## Contributing

See the [contribution guide](CONTRIBUTING.md) for details on how to contribute.
