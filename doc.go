/*
Package main is responsible for setting up the imup client
and running connection and speed testing at regular intervals.

# Default Setup

On startup any cached jobs will POST to the imup API.

The main entrypoint sets up a shutdown signal used to coordinate graceful shutdown
of the following go routines.

Authorization
Random Speed Testing
Connectivity Testing
Realtime

	Liveness
	On Demand Speed Tests
	Remote Configuration Reloading

# Advanced Configuration

See the readme for a detailed list of configurable options.
*/

package main
