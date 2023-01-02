imUp client
--

![CI](https://github.com/imup-io/client/workflows/CI/badge.svg)
![S](https://github.com/imup-io/client/workflows/CodeQL/badge.svg)

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
   go install github.com/imup-io/client
   ```

Run the client specifying an email address you want to associate the data with

  ```sh
  imup --email email@example.com
  ```

## Documentation

[pkg.go.dev](https://pkg.go.dev/github.com/imup-io/client)

## Configuration

|        Name                        |      Description                          |                   Default                                    |
|------------------------------------|-------------------------------------------|--------------------------------------------------------------|
| `ENVIRONMENT`                      | controls log output                       | `"production"`                                               |
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
| `CONN_DELAY`                       | time between dials in milliseconds        | `"200"`                                                      |
| `CONN_INTERVAL`                    | dialer interval in seconds                | `"60"`                                                       |
| `CONN_REQUESTS`                    | number of requests each test              | `"300"`                                                      |
| `SPEEDTEST_ENABLED`                | enable speed tests                        | `"false"`                                                    |
| `SPEEDTEST_INTERVAL`               | intended to set cron frequency            | VARIABLE CURRENTLY UNUSED                                    |

## Flags

```sh
Usage: imupClient
  -anonymize.ip value
     Valid values are "none" and "netblock". (default none)
  -config-version string
     config version
  -email string
     email address
  -environment string
     imUp environment (development, production)
  -id string
     host id
  -insecure
     run insecure speed tests (ws:// and not wss://)
  -key string
     api key
  -locate.url value
     The base url for the Locate API (default https://locate.measurementlab.net/v2/nearest/)
  -no-gateway-discovery
     do not attempt to discover a default gateway
  -no-speed-test
     don't run speed tests
  -non-volatile
     use disk to store collected data between tests to ensure reliability
  -ping
     use ICMP ping for connectivity tests
  -quiet-speed-test
     don't output speed test logs
  -realtime
     enable real-time features, enabled by default (default true)

```

## Contributing

Fork the repository on GitHub and clone the repository onto your machine.

Pull requests are welcome! For major changes, please open an issue first to ensure your change aligns with our overall road-map.

Please update tests and include code comments where appropriate.
