// Copyright 2022 GoEdge goedge.cdn@gmail.com. All rights reserved. Official site: https://cdn.foyeseo.com .
//go:build linux

package nftables

import (
	"fmt"
	"github.com/dashenmiren/EdgeCommon/pkg/nodeconfigs"
	teaconst "github.com/dashenmiren/EdgeNode/internal/const"
	"github.com/dashenmiren/EdgeNode/internal/events"
	"github.com/dashenmiren/EdgeNode/internal/remotelogs"
	executils "github.com/dashenmiren/EdgeNode/internal/utils/exec"
	"github.com/dashenmiren/EdgeNode/internal/utils/goman"
	"github.com/iwind/TeaGo/logs"
	"os"
	"runtime"
	"time"
)

func init() {
	if !teaconst.IsMain {
		return
	}

	events.On(events.EventReload, func() {
		// linux only
		if runtime.GOOS != "linux" {
			return
		}

		nodeConfig, err := nodeconfigs.SharedNodeConfig()
		if err != nil {
			return
		}

		if nodeConfig == nil || !nodeConfig.AutoInstallNftables {
			return
		}

		if os.Getgid() == 0 { // root user only
			if len(NftExePath()) > 0 {
				return
			}
			goman.New(func() {
				err := NewInstaller().Install()
				if err != nil {
					// 不需要传到API节点
					logs.Println("[NFTABLES]install nftables failed: " + err.Error())
				}
			})
		}
	})
}

// NftExePath 查找nftables可执行文件路径
func NftExePath() string {
	path, _ := executils.LookPath("nft")
	if len(path) > 0 {
		return path
	}

	for _, possiblePath := range []string{
		"/usr/sbin/nft",
	} {
		_, err := os.Stat(possiblePath)
		if err == nil {
			return possiblePath
		}
	}

	return ""
}

type Installer struct {
}

func NewInstaller() *Installer {
	return &Installer{}
}

func (this *Installer) Install() error {
	// linux only
	if runtime.GOOS != "linux" {
		return nil
	}

	// 检查是否已经存在
	if len(NftExePath()) > 0 {
		return nil
	}

	var cmd *executils.Cmd
	var aptCmd *executils.Cmd

	// check dnf
	{
		dnfExe, err := executils.LookPath("dnf")
		if err == nil {
			cmd = executils.NewCmd(dnfExe, "-y", "install", "nftables")
		}
	}

	// check apt
	if cmd == nil {
		aptGetExe, err := executils.LookPath("apt-get")
		if err == nil {
			cmd = executils.NewCmd(aptGetExe, "install", "nftables")
		}

		aptExe, aptErr := executils.LookPath("apt")
		if aptErr == nil {
			aptCmd = executils.NewCmd(aptExe, "install", "nftables")
		}
	}

	// check yum
	if cmd == nil {
		yumExe, err := executils.LookPath("yum")
		if err == nil {
			cmd = executils.NewCmd(yumExe, "-y", "install", "nftables")
		}
	}

	if cmd == nil {
		return nil
	}

	cmd.WithTimeout(10 * time.Minute)
	cmd.WithStderr()
	err := cmd.Run()
	if err != nil {
		// try 'apt-get' instead of 'apt'
		if aptCmd != nil {
			cmd = aptCmd
			cmd.WithTimeout(10 * time.Minute)
			cmd.WithStderr()
			err = cmd.Run()
		}

		if err != nil {
			return fmt.Errorf("%w: %s", err, cmd.Stderr())
		}
	}

	remotelogs.Println("NFTABLES", "installed nftables with command '"+cmd.String()+"' successfully")

	return nil
}
