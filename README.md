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

## Basic Usage

Download the client

   ```sh
   go install github.com/imup-io/client@latest
   ```

Run the client specifying an email address you want to associate the data with

  ```sh
  client --email email@example.com
  ```

## Contributing

See the [contribution guide](CONTRIBUTING.md) for details on how to contribute.

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

```txt
Usage: imupClient
  -allowlisted-ips string
     limit ips for speed or connectivity testing to this list, default is allow all
  -anonymize.ip value
     Valid values are "none" and "netblock". (default none)
  -api-post-connection-data string
     api endpoint for connectivity data, default is https://api.imup.io/v1/data/connectivity
  -api-post-speed-test-data string
     api endpoint for speed data, default is https://api.imup.io/v1/data/speedtest
  -blocklisted-ips string
     a list of ips for which no speed or connectivity testing will be run against, default is block none
  -config-version string
     config version for realtime reloadable configs
  -conn-delay string
     the delay between connectivity tests with a net dialer, default is 200ms
  -conn-interval string
     how often a dial test is run, default is 60s
  -conn-requests string
     the number of dials executed during a connectivity test, default is 300
  -email string
     email address associated with connectivity and speed data
  -group-id string
     an imup org users group id
  -host-id string
     the host id associated with this client
  -imup-data-length string
     the number of data points collected before sending data to the api, default is 15 data points
  -insecure
     run insecure speed tests (ws:// and not wss://), default is false
  -key string
     an api key associated with an imup organization
  -liveness-check-in-address string
     api endpoint for liveness checkins default is https://api.imup.io/v1/realtime/livenesscheckin
  -locate.url value
     The base url for the Locate API (default https://locate.measurementlab.net/v2/nearest/)
  -log-to-file
     if enabled, will log to the default root directory to use for user-specific cached data, default is false
  -no-gateway-discovery
     do not attempt to discover a default gateway, default is true
  -no-speed-test
     do not run speed tests, default is false
  -nonvolatile
     use disk to store collected data between tests to ensure reliability, default is false
  -ping
     use ICMP ping for connectivity tests, default is true (default true)
  -ping-address-internal string
     an internal gateway to differentiate between local networking issues and internet connectivity, by default imup attempts to discover your gateway
  -ping-addresses-external string
     external ip addresses imup uses to validate connectivity, defaults are 1.1.1.1,1.0.0.1,8.8.8.8,8.8.4.4
  -ping-delay string
     the delay between connectivity tests with ping, default is 100ms
  -ping-interval string
     how often a ping test is run, default is 60s
  -ping-requests string
     the number of icmp echos executed during a ping test, default is 600
  -realtime
     enable realtime features, default is true (default true)
  -realtime-authorized string
     api endpoint for imup realtime features, default is https://api.imup.io/v1/auth/realtimeAuthorized
  -realtime-config string
     api endpoint for imup realtime reloadable configuration, default is https://api.imup.io/v1/realtime/config
  -should-run-speed-test-address string
     api endpoint for imup realtime speed tests, default is https://api.imup.io/v1/realtime/shouldClientRunSpeedTest
  -speed-test-results-address string
     api endpoint for imup realtime speed test results, default is https://api.imup.io/v1/realtime/speedTestResults
  -speed-test-status-update-address string
     api endpoint for imup realtime speed tests status, default is https://api.imup.io/v1/realtime/speedTestStatusUpdate
  -verbosity string
     verbosity fo log output, default is info
```
