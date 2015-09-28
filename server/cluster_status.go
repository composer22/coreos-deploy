package server

import (
	"bufio"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

// ClusterStatus is the master header for info on the cluster.
type ClusterStatus struct {
	Machines []*ClusterMachine `json:"machines"`
}

// ClusterMachine represents each coreOS instance in the cluster.
type ClusterMachine struct {
	MachineID string         `json:"machine"`  // Machine ID
	IP        string         `json:"ip"`       // IP address of the machine
	MetaData  string         `json:"metadata"` // Metadata and sub-cluster
	Units     []*ClusterUnit `json:"units"`    // List of units running under the machine
}

// ClusterUnit represents a systemd unit being run on the machine.
type ClusterUnit struct {
	Unit   string `json:"unit"`   // The name of the unit being run.
	Hash   string `json:"hash"`   // ID of the unit being run.
	Active string `json:"active"` // Whether it's active.
	Load   string `json:"load"`   // Whether it's loaded.
	Sub    string `json:"sub"`    // Whether it's running.
}

// NewClusterStatus is a factory function that returns a new instance of ClusterStatus.
func NewClusterStatus() *ClusterStatus {
	return &ClusterStatus{
		Machines: make([]*ClusterMachine, 0),
	}
}

// NewClusterMachine is a factory function that returns a new instance of NewClusterMachine.
func NewClusterMachine(machine string, ip string, metaData string) *ClusterMachine {
	return &ClusterMachine{
		MachineID: machine,
		IP:        ip,
		MetaData:  metaData,
		Units:     make([]*ClusterUnit, 0),
	}
}

// NewClusterUnit is a factory function that returns a new instance of NewClusterUnit.
func NewClusterUnit(unit string, hash string, active string, load string, sub string) *ClusterUnit {
	return &ClusterUnit{
		Unit:   unit,
		Hash:   hash,
		Active: active,
		Load:   load,
		Sub:    sub,
	}
}

// GetClusterInfo returns a structure that represents the state of the cluster services.
func GetClusterInfo(machineQuery string, unitQuery string) (*ClusterStatus, error) {
	// Both query strings are optional.
	if machineQuery == "" {
		machineQuery = ".*"
	}
	if unitQuery == "" {
		unitQuery = ".*"
	}
	mre := regexp.MustCompile(machineQuery)
	ure := regexp.MustCompile(unitQuery)

	// Get the machines.
	cmd := exec.Command(fleetctl, "list-machines", "-fields=machine,ip,metadata", "-full=true", "-l=true", "-no-legend")
	stdout, err := execCmd(cmd)
	if err != nil {
		return nil, err
	}
	machines := make(map[string]*ClusterMachine, 0)
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		line := scanner.Text()
		if mre.FindStringIndex(line) == nil {
			continue
		}
		words := filterEmpty(strings.Split(line, "\t"))
		machines[words[0]] = NewClusterMachine(words[0], words[1], words[2])
	}
	if err := scanner.Err(); err != nil {
		return nil, scanner.Err()
	}

	// Get the units.
	cmd = exec.Command(fleetctl, "list-units", "-fields=machine,unit,hash,active,load,sub", "-full=true", "-l=true",
		"-no-legend")
	stdout, err = execCmd(cmd)
	if err != nil {
		return nil, err
	}
	scanner = bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		line := scanner.Text()
		if ure.FindStringIndex(line) == nil {
			continue
		}
		words := filterEmpty(strings.Split(line, "\t"))
		mc := strings.Split(words[0], "/")
		if _, ok := machines[mc[0]]; !ok {
			continue
		}
		machines[mc[0]].Units = append(machines[mc[0]].Units, NewClusterUnit(words[1], words[2], words[3], words[4],
			words[5]))
	}
	if err := scanner.Err(); err != nil {
		return nil, scanner.Err()
	}

	sortUnitsByUnit := func(u1, u2 *ClusterUnit) bool {
		return u1.Unit < u2.Unit
	}

	sortUnitsBySub := func(u1, u2 *ClusterUnit) bool {
		return u1.Sub < u2.Sub
	}

	sortMachinesByMetaData := func(m1, m2 *ClusterMachine) bool {
		return m1.MetaData < m2.MetaData
	}
	sortMachinesByIP := func(m1, m2 *ClusterMachine) bool {
		return m1.IP < m2.IP
	}

	s := NewClusterStatus()
	for _, m := range machines {
		// Filter out no units from machines.
		if unitQuery != ".*" && len(m.Units) <= 0 {
			continue
		}
		NewUnitSorter(sortUnitsBySub, sortUnitsByUnit).Sort(m.Units)
		s.Machines = append(s.Machines, m)
	}

	NewMachineSorter(sortMachinesByMetaData, sortMachinesByIP).Sort(s.Machines)
	return s, nil
}

// filterEmpty returns an array containing all non-empty strings of the input array.
func filterEmpty(values []string) []string {
	result := make([]string, 0)
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			result = append(result, v)
		}
	}
	return result
}

type machineLessFunc func(p1, p2 *ClusterMachine) bool

// MachineSorter implements the Sort interface, sorting the changes within.
type MachineSorter struct {
	Machines []*ClusterMachine
	less     []machineLessFunc
}

// NewMachineSorter returns a Sorter that sorts using the less functions, in order.
func NewMachineSorter(less ...machineLessFunc) *MachineSorter {
	return &MachineSorter{
		less: less,
	}
}

// Sort sorts the argument slice according to the less functions passed to OrderedBy.
func (ms *MachineSorter) Sort(machines []*ClusterMachine) {
	ms.Machines = machines
	sort.Sort(ms)
}

// Len is part of sort.Interface.
func (ms *MachineSorter) Len() int {
	return len(ms.Machines)
}

// Swap is part of sort.Interface.
func (ms *MachineSorter) Swap(i, j int) {
	ms.Machines[i], ms.Machines[j] = ms.Machines[j], ms.Machines[i]
}

// Less is part of sort.Interface.
func (ms *MachineSorter) Less(i, j int) bool {
	p, q := ms.Machines[i], ms.Machines[j]
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}
	return ms.less[k](p, q)
}

type unitLessFunc func(p1, p2 *ClusterUnit) bool

// UnitSorter implements the Sort interface, sorting the changes within.
type UnitSorter struct {
	Units []*ClusterUnit
	less  []unitLessFunc
}

// NewMachineSorter returns a Sorter that sorts using the less functions, in order.
func NewUnitSorter(less ...unitLessFunc) *UnitSorter {
	return &UnitSorter{
		less: less,
	}
}

// Sort sorts the argument slice according to the less functions passed to OrderedBy.
func (us *UnitSorter) Sort(units []*ClusterUnit) {
	us.Units = units
	sort.Sort(us)
}

// Len is part of sort.Interface.
func (us *UnitSorter) Len() int {
	return len(us.Units)
}

// Swap is part of sort.Interface.
func (us *UnitSorter) Swap(i, j int) {
	us.Units[i], us.Units[j] = us.Units[j], us.Units[i]
}

// Less is part of sort.Interface.
func (us *UnitSorter) Less(i, j int) bool {
	p, q := us.Units[i], us.Units[j]
	var k int
	for k = 0; k < len(us.less)-1; k++ {
		less := us.less[k]
		switch {
		case less(p, q):
			return true
		case less(q, p):
			return false
		}
	}
	return us.less[k](p, q)
}
