//go:build darwin && cgo

package main

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"
)

//go:embed module_network_wake.sh
var moduleNetworkWakeScript []byte

//go:embed module_network_wake_init.sh
var moduleNetworkWakeInitScript []byte

const (
	moduleNetworkWakeBinary = "/data/djonehub/bin/djonehub-network-wake"
	moduleNetworkWakeInit   = "/etc/init.d/djonehub_network_wake"
	moduleNetworkWakeLink   = "/etc/rc5.d/S98djonehub_network_wake"
	moduleNetworkWakeFlag   = "/data/djonehub/network-wake.enabled"
)

func enableModuleNetworkWake() error {
	adb, err := openDJIUSBADB()
	if err != nil {
		return fmt.Errorf("打开模块 ADB 失败: %w", err)
	}
	defer adb.Close()
	if err := requireModuleRoot(adb); err != nil {
		return err
	}
	if err := adb.push(moduleNetworkWakeScript, "/data/local/tmp/djonehub-network-wake.new", 0o100755, 15*time.Second); err != nil {
		return fmt.Errorf("上传网络唤醒服务失败: %w", err)
	}
	if err := adb.push(moduleNetworkWakeInitScript, "/data/local/tmp/djonehub_network_wake.new", 0o100755, 15*time.Second); err != nil {
		return fmt.Errorf("上传网络唤醒启动项失败: %w", err)
	}
	command := "set -e; mkdir -p /data/djonehub/bin /data/djonehub/log /data/djonehub/backup; " +
		"stamp=$(date +%Y%m%d-%H%M%S); backup=/data/djonehub/backup/network-wake-$stamp; mkdir -p $backup; " +
		"test ! -e " + moduleNetworkWakeBinary + " || cp -p " + moduleNetworkWakeBinary + " $backup/; " +
		"test ! -e " + moduleNetworkWakeInit + " || cp -p " + moduleNetworkWakeInit + " $backup/; " +
		"cp /data/local/tmp/djonehub-network-wake.new " + moduleNetworkWakeBinary + "; chmod 755 " + moduleNetworkWakeBinary + "; " +
		"cp /data/local/tmp/djonehub_network_wake.new " + moduleNetworkWakeInit + "; chmod 755 " + moduleNetworkWakeInit + "; " +
		"ln -sfn ../init.d/djonehub_network_wake " + moduleNetworkWakeLink + "; " +
		"touch " + moduleNetworkWakeFlag + "; " + moduleNetworkWakeInit + " restart; " + moduleNetworkWakeInit + " status"
	out, status, err := adb.shellChecked(command, 25*time.Second)
	if err != nil {
		return fmt.Errorf("安装网络唤醒服务失败: %w", err)
	}
	if status != 0 {
		return fmt.Errorf("网络唤醒服务未启动（status=%d, output=%s）", status, strings.TrimSpace(out))
	}
	return nil
}

func disableModuleNetworkWake() error {
	adb, err := openDJIUSBADB()
	if err != nil {
		return err
	}
	defer adb.Close()
	if err := requireModuleRoot(adb); err != nil {
		return err
	}
	out, status, err := adb.shellChecked("rm -f "+moduleNetworkWakeFlag+"; test ! -x "+moduleNetworkWakeInit+" || "+moduleNetworkWakeInit+" stop", 12*time.Second)
	if err != nil {
		return err
	}
	if status != 0 {
		return fmt.Errorf("停用网络唤醒服务失败（status=%d, output=%s）", status, strings.TrimSpace(out))
	}
	return nil
}

func requireModuleRoot(adb *adbClient) error {
	out, status, err := adb.shellChecked("id -u", 8*time.Second)
	if err != nil {
		return err
	}
	isRoot := false
	for _, field := range strings.Fields(out) {
		if field == "0" {
			isRoot = true
			break
		}
	}
	if status != 0 || !isRoot {
		return errors.New("模块 ADB 没有 root 权限，无法安装网络唤醒服务")
	}
	return nil
}
