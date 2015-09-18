package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/composer22/coreos-deploy/db"
	"github.com/composer22/coreos-deploy/etcd2"
)

const (
	etc2CurrentCycleTmpl = "/%s/apps/services/%s/current-cycle"
	etc2CurrentUnitTmpl  = "/%s/apps/services/%s/current-cycle-unit"
	etc2CurrentCountTmpl = "/%s/apps/services/%s/current-cycle-count"
)

// ServiceRequest is a struct used to demarshal requests for a deploy.
type ServiceRequest struct {
	ServiceName     string              `json:"serviceName"`     // The name of the service to deploy.
	Version         string              `json:"version"`         // The version of the deploy.
	NumInstances    int                 `json:"numInstances"`    // The number of instances to deploy.
	ServiceTemplate string              `json:"serviceTemplate"` // Source code for the unit template.
	Etcd2Keys       map[string]string   `json:"etcd2Keys"`       // etcd2 keys to update.
	Suffix          string              `json:"-"`               // A unique suffix for the new service.
	Domain          string              `json:"-"`               // What domain this cluster is serving.
	Environment     string              `json:"-"`               // The environment (dev, stage, prod, etc).
	DeployID        string              `json:"-"`               // A UUID for the request and for this deploy.
	mu              *sync.RWMutex       `json:"-"`               // One deploy at a time for this server.
	wg              *sync.WaitGroup     `json:"-"`               // The wait group.
	db              *db.DBConnect       `json:"-"`               // The DB connection for status updates.
	e2              *etcd2.Etcd2Connect `json:"-"`               // The etcd2 connection point.
}

// NewServiceRequest is a factory function that returns a ServiceRequest instance.
func NewServiceRequest(name string, vers string, instances int, template string, keys map[string]string) *ServiceRequest {
	return &ServiceRequest{
		ServiceName:     name,
		Version:         vers,
		NumInstances:    instances,
		ServiceTemplate: template,
		Etcd2Keys:       keys,
	}
}

// Deploy is a go routine that attempts to update etcd2 and/or run fleetctl to start a service in coreOS.
func (r *ServiceRequest) Deploy() {
	defer r.wg.Done()
	r.mu.Lock()
	defer r.mu.Unlock()

	var log string = ""

	// Write the start of job record to the DB.
	r.db.StartDeploy(r.DeployID, r.Domain, r.Environment, r.ServiceName, r.Version, r.NumInstances,
		r.ServiceTemplate, r.Etcd2Keys, r.Suffix)

	// Save service unit code.
	log += "Saving service unit code to temp file.\n"
	serviceFileName := fmt.Sprintf("%s-%s-%s@.service", r.ServiceName, r.Version, r.Suffix)
	serviceFilePath := fmt.Sprintf("%s%s", tmpDir, serviceFileName)
	err := ioutil.WriteFile(serviceFilePath, []byte(r.ServiceTemplate), 0644)
	if err != nil {
		msg := "Unable to write service unit file to temp."
		log += fmt.Sprintf("ERR: %s\n%s\n", msg, err)
		r.db.UpdateDeploy(r.DeployID, db.Failed, msg, log)
		return
	}

	// Apply etcd2 key changes.
	log += "Applying etcd2 key changes.\n"
	if err := r.e2.Set(r.Etcd2Keys); err != nil {
		msg := "Unable to apply etcd2 key changes."
		log += fmt.Sprintf("ERR: %s\n%s\n", msg, err)
		r.db.UpdateDeploy(r.DeployID, db.Failed, msg, log)
		return
	}

	// Install service template.
	log += "Install service template.\n"
	cmd := exec.Command(fleetctl, "destroy", serviceFileName)
	if _, err := execCmd(cmd); err != nil {
		msg := err.Error()
		if msg != "exit status 1" && !strings.Contains(msg, "unit does not exist") {
			msg = "Unable to destroy previous service for new template."
			log += fmt.Sprintf("ERR: %s\n%s\n", msg, err)
			r.db.UpdateDeploy(r.DeployID, db.Failed, msg, log)
			return
		}
	}

	cmd = exec.Command(fleetctl, "submit", serviceFilePath)
	if _, err := execCmd(cmd); err != nil {
		msg := "Unable to submit service template."
		log += fmt.Sprintf("ERR: %s\n%s\n", msg, err)
		r.db.UpdateDeploy(r.DeployID, db.Failed, msg, log)
		return
	}

	log += "Deleting temp unit file.\n"
	os.Remove(serviceFilePath)

	// Start new services in the cluster.
	log += "Performing A/B rotation of service.\n"
	if err := r.flipAB(); err != nil {
		msg := "Unable to perform A/B rotation of service."
		log += fmt.Sprintf("ERR: %s\n%s\n", msg, err)
		r.db.UpdateDeploy(r.DeployID, db.Failed, msg, log)
		return
	}

	// Update the job record with a success.
	msg := "Service deployed successfully."
	log += fmt.Sprintf("SUCCESS: %s\n", msg)
	r.db.UpdateDeploy(r.DeployID, db.Success, msg, log)
}

// flipAB instantiates new instance using the service template and takes down previous services.
func (r *ServiceRequest) flipAB() error {
	// Initialize keys
	currentCycleKey := fmt.Sprintf(etc2CurrentCycleTmpl, r.Domain, r.ServiceName)
	currentUnitKey := fmt.Sprintf(etc2CurrentUnitTmpl, r.Domain, r.ServiceName)
	currentCountKey := fmt.Sprintf(etc2CurrentCountTmpl, r.Domain, r.ServiceName)

	// Set the default values for first deploy to the system for this application.
	etc2Keys := make(map[string]string)
	etc2Keys[currentCycleKey] = "B"
	etc2Keys[currentUnitKey] = "*coreos-deploy-noop"
	etc2Keys[currentCountKey] = "0"
	r.e2.Make(etc2Keys)

	// Get current cycle info.
	etc2Keys, err := r.e2.Get(etc2Keys)
	if err != nil {
		return err
	}

	// Initialize new cycle indicator based on current_cycle.
	newCycle := "A"
	if etc2Keys[currentCycleKey] == "A" {
		newCycle = "B"
	}

	// Start n new instances in the cluster.
	for i := 1; i <= r.NumInstances; i++ {
		serviceCmd := fmt.Sprintf("%s-%s-%s@%s%d.service", r.ServiceName, r.Version, r.Suffix, newCycle, i)
		cmd := exec.Command(fleetctl, "stop", serviceCmd)
		execCmd(cmd)
		cmd = exec.Command(fleetctl, "destroy", serviceCmd)
		execCmd(cmd)
		cmd = exec.Command(fleetctl, "start", serviceCmd)
		if _, err := execCmd(cmd); err != nil {
			return err
		}
	}

	// Take down old services.
	if etc2Keys[currentUnitKey] != "*coreos-deploy-noop" {
		cc, _ := strconv.Atoi(etc2Keys[currentCountKey])
		for i := 1; i <= cc; i++ {
			serviceCmd := fmt.Sprintf("%s@%s%d.service", etc2Keys[currentUnitKey], etc2Keys[currentCycleKey], i)
			fmt.Println(serviceCmd)
			cmd := exec.Command(fleetctl, "stop", serviceCmd)
			execCmd(cmd)
			cmd = exec.Command(fleetctl, "destroy", serviceCmd)
			execCmd(cmd)
		}
		// Destroy old template.
		cmd := exec.Command(fleetctl, "destroy", fmt.Sprintf("%s@.service", etc2Keys[currentUnitKey]))
		execCmd(cmd)
	}

	// Set current cycle to new values for next time.
	etc2Keys[currentCycleKey] = newCycle
	etc2Keys[currentUnitKey] = fmt.Sprintf("%s-%s-%s", r.ServiceName, r.Version, r.Suffix)
	etc2Keys[currentCountKey] = strconv.Itoa(r.NumInstances)
	if err := r.e2.Set(etc2Keys); err != nil {
		return err
	}
	return nil
}
