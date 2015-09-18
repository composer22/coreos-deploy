package db

import (
	"database/sql"
	"encoding/json"

	_ "github.com/go-sql-driver/mysql"
)

const (
	_ = iota
	Started
	Success
	Failed
)

type DBConnect struct {
	db *sql.DB
}

// NewDBConnect is a factory method that returns a new db connection
func NewDBConnect(dsn string) (*DBConnect, error) {

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Validate DSN data
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DBConnect{db: db}, nil
}

// ValidAuth returns true if the API Key is valid for a request.
func (d *DBConnect) ValidAuth(key string) bool {
	var id int
	row := d.db.QueryRow("SELECT id FROM auth_tokens WHERE token = ?", key)
	err := row.Scan(&id)

	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		return false
	default:
		return true
	}
}

// StartDeploy inserts a fresh row into the log for a deployment run.
func (d *DBConnect) StartDeploy(deployID string, domain string, environment string, serviceName string,
	version string, numInstances int, serviceTemplate string, etcd2Keys map[string]string,
	suffix string) bool {
	etcd2, _ := json.Marshal(etcd2Keys)
	result, err := d.db.Exec("INSERT INTO deploys (deploy_id, domain, environment, service_name, version, "+
		"num_instances, service_template, etcd2_keys, status, suffix, message, log, updated_at, created_at) "+
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, \"Start deploy.\", \"Start deploy.\", NOW(), NOW())",
		deployID, domain, environment, serviceName, version, numInstances, serviceTemplate, etcd2, Started, suffix)
	if err != nil {
		return false
	}
	id, err := result.LastInsertId()
	if err != nil || id <= 0 {
		return false
	}
	return true
}

// UpdateDeploy updates the deploy row with information from the run.
func (d *DBConnect) UpdateDeploy(deployID string, status int, message string, log string) bool {
	result, err := d.db.Exec("UPDATE deploys "+
		"SET status = ?, "+
		"message = ?, "+
		"log = ?, "+
		"updated_at = NOW() "+
		"WHERE deploy_id = ?",
		status, message, log, deployID)
	if err != nil {
		return false
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return false
	}
	return true
}

// DeployStatus is used to return deploy status information from the database to the requester.
type DeployStatus struct {
	DeployID     string `json:"deployID"`     // The deploy UUID.
	Domain       string `json:"domain"`       // The domain name serviced.
	Environment  string `json:"environment"`  // The environment serviced (development, qa etc.)
	ServiceName  string `json:"serviceName"`  // The application name of the service ex: video-mobile.
	Version      string `json:"version"`      // The version of teh application ex; 1.0.0
	Suffix       string `json:"suffix"`       // The suffix added to the service name.
	NumInstances int    `json:"numInstances"` // The number of instances deployed.
	Status       int    `json:"status"`       // The status ID of the result.
	Message      string `json:"message"`      // A user friendly message of what occurred.
	Log          string `json:"log"`          // The log of all steps run during the deploy.
	UpdatedAt    string `json:"updatedAt"`    // The create date and time of the deploy.
	CreatedAt    string `json:"createdAt"`    // The last update to this record.
}

// QueryDeploy returns the status of a deploy request.
func (d *DBConnect) QueryDeploy(deployID string) (*DeployStatus, error) {
	r := &DeployStatus{}
	row := d.db.QueryRow("SELECT deploy_id, domain, environment, service_name, version, "+
		"num_instances, status, suffix, message, log, updated_at, created_at "+
		"FROM deploys WHERE deploy_id = ?", deployID)
	err := row.Scan(&r.DeployID, &r.Domain, &r.Environment, &r.ServiceName, &r.Version, &r.NumInstances,
		&r.Status, &r.Suffix, &r.Message, &r.Log, &r.UpdatedAt, &r.CreatedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil, err
	case err != nil:
		return nil, err
	default:
		return r, nil
	}
}

// Close closes the connection(s) to the DB.
func (d *DBConnect) Close() bool {
	d.db.Close()
	return true
}
