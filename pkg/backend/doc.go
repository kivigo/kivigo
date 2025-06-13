/*
Package backend provides backend connectors for KiviGo.

A backend implements the storage logic for key-value pairs. Built-in backends include:
  - Local (BoltDB)
  - Redis

You can implement your own backend by satisfying the models.KV interface.
*/
package backend
