/*

Package config provides a thread safe, remotely reloadable base configuration
for running the imup client.

# Synchronization

This package makes use of a global read-write mutex to enable a safe update
to the Reloadable interface.  The interface is a collection of thread safe getter methods,
intentionally preventing the application to mutate state.

There are two exposed functions that enables/disables this feature.
Those functions are EnableRealtime and DisableRealtime.

The Reload function takes a payload that should correspond to a base config object
and safely mutates the global configuration

*/

package config
