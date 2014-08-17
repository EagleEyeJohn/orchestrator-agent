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

package http

import (
	"os"
	"encoding/json"
	"net/http"
	"errors"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	
	"github.com/outbrain/orchestrator-agent/config"
	"github.com/outbrain/orchestrator-agent/agent"
	"github.com/outbrain/orchestrator-agent/osagent"	
)

type HttpAPI struct{}

var API HttpAPI = HttpAPI{}


// APIResponseCode is an OK/ERROR response code
type APIResponseCode int

const (
	ERROR APIResponseCode = iota
	OK
)

func (this *APIResponseCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.String())
}

func (this *APIResponseCode) String() string {
	switch *this {
		case ERROR: return "ERROR"
		case OK: return "OK"
	}
	return "unknown"
}


// APIResponse is a response returned as JSON to various requests.
type APIResponse struct {
	Code	APIResponseCode
	Message	string
	Details	interface{}
}


func validateToken(token string) error {
	if token == agent.ProcessToken.Hash { 
		return nil
	} else {
		return errors.New("Invalid token")
	}
}

// Hostname provides information on this process
func (this *HttpAPI) Hostname(params martini.Params, r render.Render) {
	hostname, err := os.Hostname()
	if err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	r.JSON(200, hostname)
}

// ListLogicalVolumes lists logical volumes by pattern
func (this *HttpAPI) ListLogicalVolumes(params martini.Params, r render.Render, req *http.Request) {
	if err := validateToken(req.URL.Query().Get("token")); err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	output, err := osagent.LogicalVolumes("", params["pattern"])
	if err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	r.JSON(200, output)
}


// ListSnapshotsLogicalVolumes lists logical volumes by pattern
func (this *HttpAPI) ListSnapshotsLogicalVolumes(params martini.Params, r render.Render, req *http.Request) {
	if err := validateToken(req.URL.Query().Get("token")); err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	output, err := osagent.LogicalVolumes("", config.Config.SnapshotVolumesFilter)
	if err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	r.JSON(200, output)
}


// LogicalVolume lists a logical volume by name/path/mount point
func (this *HttpAPI) LogicalVolume(params martini.Params, r render.Render, req *http.Request) {
	if err := validateToken(req.URL.Query().Get("token")); err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	lv := params["lv"]
	if lv == "" {
		lv = req.URL.Query().Get("lv");
	}
	output, err := osagent.LogicalVolumes(lv, "")
	if err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	r.JSON(200, output)
}


// GetMount shows the configured mount point's status
func (this *HttpAPI) GetMount(params martini.Params, r render.Render, req *http.Request) {
	if err := validateToken(req.URL.Query().Get("token")); err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	output, err := osagent.GetMount(config.Config.SnapshotMountPoint)
	if err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	r.JSON(200, output)
}


// MountLV mounts a logical volume on config mount point
func (this *HttpAPI) MountLV(params martini.Params, r render.Render, req *http.Request) {
	if err := validateToken(req.URL.Query().Get("token")); err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	lv := params["lv"]
	if lv == "" {
		lv = req.URL.Query().Get("lv");
	}
	output, err := osagent.MountLV(config.Config.SnapshotMountPoint, lv)
	if err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	r.JSON(200, output)
}



// Unmount umounts the config mount point
func (this *HttpAPI) Unmount(params martini.Params, r render.Render, req *http.Request) {
	if err := validateToken(req.URL.Query().Get("token")); err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	output, err := osagent.Unmount(config.Config.SnapshotMountPoint)
	if err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	r.JSON(200, output)
}



// DiskUsage returns the number of bytes of a give ndirectory (recursive)
func (this *HttpAPI) DiskUsage(params martini.Params, r render.Render, req *http.Request) {
	if err := validateToken(req.URL.Query().Get("token")); err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	path := req.URL.Query().Get("path");
	
	output, err := osagent.DiskUsage(path)
	if err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	r.JSON(200, output)
}



// LocalSnapshots lists dc-local available snapshots for this host
func (this *HttpAPI) AvailableLocalSnapshots(params martini.Params, r render.Render, req *http.Request) {
	if err := validateToken(req.URL.Query().Get("token")); err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	output, err := osagent.AvailableSnapshots(true)
	if err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	r.JSON(200, output)
}


// Snapshots lists available snapshots for this host
func (this *HttpAPI) AvailableSnapshots(params martini.Params, r render.Render, req *http.Request) {
	if err := validateToken(req.URL.Query().Get("token")); err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	output, err := osagent.AvailableSnapshots(false)
	if err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	r.JSON(200, output)
}


// MySQLRunning checks whether the MySQL service is up
func (this *HttpAPI) MySQLRunning(params martini.Params, r render.Render, req *http.Request) {
	if err := validateToken(req.URL.Query().Get("token")); err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	output, err := osagent.MySQLRunning()
	if err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	r.JSON(200, output)
}


// MySQLStop shuts down the MySQL service
func (this *HttpAPI) MySQLStop(params martini.Params, r render.Render, req *http.Request) {
	if err := validateToken(req.URL.Query().Get("token")); err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	err := osagent.MySQLStop()
	if err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	r.JSON(200, err == nil)
}



// MySQLStop starts the MySQL service
func (this *HttpAPI) MySQLStart(params martini.Params, r render.Render, req *http.Request) {
	if err := validateToken(req.URL.Query().Get("token")); err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	err := osagent.MySQLStart()
	if err != nil {
		r.JSON(500, &APIResponse{Code:ERROR, Message: err.Error(),})
		return
	}
	r.JSON(200, err == nil)
}



// RegisterRequests makes for the de-facto list of known API calls
func (this *HttpAPI) RegisterRequests(m *martini.ClassicMartini) {
	m.Get("/api/hostname", this.Hostname) 
	m.Get("/api/lvs", this.ListLogicalVolumes) 
	m.Get("/api/lvs/:pattern", this.ListLogicalVolumes) 
	m.Get("/api/lvs-snapshots", this.ListSnapshotsLogicalVolumes) 
	m.Get("/api/lv", this.LogicalVolume) 
	m.Get("/api/lv/:lv", this.LogicalVolume) 
	m.Get("/api/mount", this.GetMount) 
	m.Get("/api/mountlv", this.MountLV) 
	m.Get("/api/umount", this.Unmount) 
	m.Get("/api/du", this.DiskUsage) 
	m.Get("/api/available-snapshots-local", this.AvailableLocalSnapshots) 
	m.Get("/api/available-snapshots", this.AvailableSnapshots) 
	m.Get("/api/mysql-status", this.MySQLRunning) 
	m.Get("/api/mysql-stop", this.MySQLStop) 
	m.Get("/api/mysql-start", this.MySQLStart) 
}