package etcd2

import (
	"fmt"
	"time"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

// Etcd2Connect represents a connection to the etcd2 server.
type Etcd2Connect struct {
	etcd2 client.Client
}

// NewDBConnect is a factory method that returns a new db connection.
func NewEtcd2Connect(etcd2EndPoint string) (*Etcd2Connect, error) {
	c, err := client.New(client.Config{
		Endpoints:               []string{fmt.Sprintf("http://%s", etcd2EndPoint)},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &Etcd2Connect{etcd2: c}, nil
}

// Set sets the etcd2 key with a value and returns the response or an error.
func (e *Etcd2Connect) Set(data map[string]string) error {
	kapi := client.NewKeysAPI(e.etcd2)
	for k, v := range data {
		if _, err := kapi.Set(context.Background(), k, v, nil); err != nil {
			return err
		}
	}
	return nil
}

// Make creates the etcd2 keys if they do not exist
func (e *Etcd2Connect) Make(data map[string]string) error {
	kapi := client.NewKeysAPI(e.etcd2)
	opts := &client.SetOptions{PrevExist: client.PrevNoExist}
	for k, v := range data {
		if _, err := kapi.Set(context.Background(), k, v, opts); err != nil {
			return err
		}
	}
	return nil
}

// Get returns the etcd2 keys for a given set of keys
func (e *Etcd2Connect) Get(data map[string]string) (map[string]string, error) {
	kapi := client.NewKeysAPI(e.etcd2)
	result := make(map[string]string)
	for k := range data {
		resp, err := kapi.Get(context.Background(), k, nil)
		if err != nil {
			return nil, err
		}
		result[k] = resp.Node.Value
	}
	return result, nil
}
