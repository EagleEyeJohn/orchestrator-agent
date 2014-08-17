/*
   Copyright 2014 Outbrain Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package osagent

import (
	"fmt"
	"errors"
	"os"
	"os/exec"
	"strings"
	"strconv"
	"regexp"
	"github.com/outbrain/log"

	"github.com/outbrain/orchestrator-agent/config"
)


// LogicalVolume describes an LVM volume
type LogicalVolume struct {
	Name			string
	GroupName		string
	Path			string
	IsSnapshot		bool
	SnapshotPercent	float64
}


// Equals tests equality of this corrdinate and another one.
func (this *LogicalVolume) IsSnapshotValid() bool {
	if !this.IsSnapshot { return false }
	if this.SnapshotPercent >= 100.0 { return false }
	return true
}


// Mount describes a file system mount point
type Mount struct {
	Path			string
	Device			string
	LVPath			string
	FileSystem		string
	IsMounted		bool
	DiskUsage		int64
}


func commandSplit(commandText string) (string, []string) {
	tokens := regexp.MustCompile(`[ ]+`).Split(strings.TrimSpace(commandText), -1)
	return tokens[0], tokens[1:]
}


// commandOutput executes a command and return output bytes
func commandOutput(commandText string) ([]byte, error) {
	commandName, commandArgs := commandSplit(commandText)
	
	log.Debugf("commandOutput: %s", commandText)
	outputBytes, err := exec.Command(commandName, commandArgs...).Output()
	if err != nil {
		return nil, log.Errore(err)
	}
	return outputBytes, nil
}

func outputLines(commandOutput []byte, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}
	text := strings.Trim(fmt.Sprintf("%s", commandOutput), "\n")
	lines := strings.Split(text, "\n")
	return lines, err
}

func outputTokens(delimiterPattern string, commandOutput []byte, err error) ([][]string, error) {
	lines, err := outputLines(commandOutput, err)
	if err != nil {
		return nil, err
	}
	tokens := make([][]string, len(lines))
	for i := range tokens {
		tokens[i] = regexp.MustCompile(delimiterPattern).Split(lines[i], -1)
	}
	return tokens, err
}


func Hostname() (string, error) {
	return os.Hostname()
}


func LogicalVolumes(volumeName string, filterPattern string) ([]LogicalVolume, error) {
	output, err := commandOutput(fmt.Sprintf("lvs --noheading -o lv_name,vg_name,lv_path,snap_percent %s", volumeName))
	tokens, err := outputTokens(`[ \t]+`, output, err)
	if err != nil { return nil, err }
	
	logicalVolumes := []LogicalVolume{}
	for _, lineTokens := range tokens {
		logicalVolume := LogicalVolume {
			Name:		lineTokens[1],
			GroupName: 	lineTokens[2],
			Path:		lineTokens[3],
		}
		logicalVolume.SnapshotPercent, err = strconv.ParseFloat(lineTokens[4], 32)
		logicalVolume.IsSnapshot = (err == nil)
		if strings.Contains(logicalVolume.Name, filterPattern) {
	    	logicalVolumes = append(logicalVolumes, logicalVolume)
	    }
	}
	return logicalVolumes, nil
}


func GetLogicalVolumePath(volumeName string) (string, error) {
	if logicalVolumes, err := LogicalVolumes(volumeName, ""); err == nil && len(logicalVolumes) > 0 {
		return logicalVolumes[0].Path, err
	}
	return "", errors.New(fmt.Sprintf("logical volume not found: %+v", volumeName))
}


func GetMount(mountPoint string) (Mount, error) {
	mount := Mount {
		Path: mountPoint,	
		IsMounted: false,	
	}

	output, err := commandOutput(fmt.Sprintf("grep %s /etc/mtab", mountPoint))
	tokens, err := outputTokens(`[ \t]+`, output, err)
	if err != nil {
		// when grep does not find rows, it returns an error. So this is actually OK 
		return mount, nil
	}
	
	for _, lineTokens := range tokens {
		mount.IsMounted = true
		mount.Device = lineTokens[0]
		mount.Path = lineTokens[1]
		mount.FileSystem = lineTokens[2]
		mount.LVPath, _ = GetLogicalVolumePath(mount.Device)
		mount.DiskUsage, _ = DiskUsage(mountPoint)
	}
	return mount, nil
}

func MountLV(mountPoint string, volumeName string) (Mount, error) {
	mount := Mount {
		Path: mountPoint,	
		IsMounted: false,	
	}
	if volumeName == "" {
		return mount, errors.New("Empty columeName in MountLV")
	}
	_, err := commandOutput(fmt.Sprintf("mount %s %s", volumeName, mountPoint))
	if err != nil {
		return mount, err
	}
	return GetMount(mountPoint)
}

func Unmount(mountPoint string) (Mount, error) {
	mount := Mount {
		Path: mountPoint,	
		IsMounted: false,	
	}
	_, err := commandOutput(fmt.Sprintf("umount %s", mountPoint))
	if err != nil {
		return mount, err
	}
	return GetMount(mountPoint)
}


func DiskUsage(path string) (int64, error) {
	var result int64

	output, err := commandOutput(fmt.Sprintf("du -sb %s", path))
	tokens, err := outputTokens(`[ \t]+`, output, err)
	if err != nil {
		return result, err
	}
	
	for _, lineTokens := range tokens {
		result, err = strconv.ParseInt(lineTokens[0], 10, 0)
		return result, err
	}
	return result, err
}


func AvailableSnapshots(requireLocal bool) ([]string, error) {
	var command string
	if requireLocal { 
		command = config.Config.AvailableLocalSnapshotHostsCommand 
	} else {
		command = config.Config.AvailableSnapshotHostsCommand
	}
	output, err := commandOutput(command)
	hosts, err := outputLines(output, err)

	return hosts, err
}


func MySQLRunning() (bool, error) {
	_, err := commandOutput(config.Config.MySQLServiceStatusCommand)
	// status command exits with 0 when MySQL is running, or otherwise if not running
	return err == nil, nil
}

func MySQLStop() error {
	_, err := commandOutput(config.Config.MySQLServiceStopCommand)
	return err
}

func MySQLStart() error {
	_, err := commandOutput(config.Config.MySQLServiceStartCommand)
	return err
}

