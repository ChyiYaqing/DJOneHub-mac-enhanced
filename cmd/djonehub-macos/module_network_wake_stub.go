//go:build !darwin || !cgo

package main

func enableModuleNetworkWake() error  { return nil }
func disableModuleNetworkWake() error { return nil }
