package etcsrc

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/gliderlabs/logspout/router"
	"golang.org/x/net/context"
)

func init() {
	src := &EtcdSource{
		hostkey: getopt("LOGGER_HOST_KEY", "/deis/logs/host"),
		portkey: getopt("LOGGER_PORT_KEY", "/deis/logs/port"),
	}
	router.Jobs.Register(src, "etcsrc")
}

func getopt(name, dfault string) string {
    value := os.Getenv(name)
    if value == "" {
        value = dfault
    }
    return value
}

type EtcdSource struct {
	hostkey string
	portkey string
	kapi    client.KeysAPI
}

func (p *EtcdSource) Name() string {
	return "etcsrc"
}

func (p *EtcdSource) Setup() error {
	url := os.Getenv("ETCD_URL")
	if url == "" {
		host := os.Getenv("ETCD_HOST")
		if host == "" {
			return errors.New("ETCD_URL/ETCD_HOST not defined")
		}
		url = fmt.Sprintf("http://%v:4001", host)
	}
	cfg := client.Config{
		Endpoints: []string{url},
		Transport: client.DefaultTransport,
	}
	c, err := client.New(cfg)
	if err != nil {
		return err
	}
	kapi := client.NewKeysAPI(c)
	p.kapi = kapi
	return nil
}

func (p *EtcdSource) Run() error {
	keysThere := false

	var route *router.Route
	var err error

	for !keysThere {
		route, err = p.processroute(nil)
		if err != nil {
			etcdErr, isEtcdErr := err.(client.Error)
			if !isEtcdErr {
				return err
			}
			if etcdErr.Code != client.ErrorCodeKeyNotFound {
				return err
			}
			// Sleep for 5 secodns if the keys aren't there, then try again.
			log.Printf("Keys %v and %v not found.  Sleeping", p.hostkey, p.portkey)
			time.Sleep(5000 * time.Millisecond)
		} else {
			keysThere = true
		}
	}

	watcher := p.kapi.Watcher(p.hostkey, nil)
	for true {
		_, err := watcher.Next(context.Background())
		if err != nil {
			return err
		}
		log.Print("Etcd watcher returned.  Re-checking route.")
		route, err = p.processroute(route)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *EtcdSource) processroute(current *router.Route) (*router.Route, error) {
	hostNode, err := p.kapi.Get(context.Background(), p.hostkey, nil)
	if err != nil {
		return nil, err
	}
	host := hostNode.Node.Value

	portNode, err := p.kapi.Get(context.Background(), p.portkey, nil)
	if err != nil {
		return nil, err
	}
	port := portNode.Node.Value

	new := &router.Route{
		Address: fmt.Sprintf("%v:%v", host, port),
		Adapter: "syslog",
	}
	if current != nil && new.Address == current.Address {
		log.Print("No address changes")
		return current, nil
	}
	log.Print("Address has changed.  Adding old & removing new route.")

	route_err := router.Routes.Add(new)
	if route_err != nil {
		return nil, route_err
	}
	if current != nil {
		router.Routes.Remove(current.ID)
	}

	return new, nil
}
