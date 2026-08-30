//go:build darwin && cgo

package main

import (
	"bytes"
	"testing"
)

func TestModuleNetworkWakeIsMobileScopedAndReleasesLock(t *testing.T) {
	checks := [][]byte{
		[]byte("network-wake.enabled"),
		[]byte("CONFIGURED"),
		[]byte("*,ecm,*"),
		[]byte("*,audio,*"),
		[]byte("wake_lock"),
		[]byte("wake_unlock"),
		[]byte("192.168.225.2"),
		[]byte("1.1.1.1"),
	}
	for _, check := range checks {
		if !bytes.Contains(moduleNetworkWakeScript, check) {
			t.Fatalf("network wake script missing %q", check)
		}
	}
}

func TestModuleNetworkWakeInitHasLifecycleCommands(t *testing.T) {
	for _, check := range [][]byte{[]byte("start)"), []byte("stop)"), []byte("restart)"), []byte("status)")} {
		if !bytes.Contains(moduleNetworkWakeInitScript, check) {
			t.Fatalf("network wake init missing %q", check)
		}
	}
}
