package telescope

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bluebreezecf/opentsdb-goclient/client"
	opentsdbConf "github.com/bluebreezecf/opentsdb-goclient/config"
)

const (
	AppID     = "APP_ID"
	TsdbHost  = "TSDB"
	Interval  = 10 * time.Second
	BatchSize = 20
)

var (
	telescope *Telescope
)

func init() {
	telescope = NewTelescope()
}

// config documents here
type config struct {
	AppName, Group, Host string
}

func envConfig() (*config, error) {
	id := os.Getenv(AppID)
	if id == "" {
		return nil, errors.New("No APP_ID found")
	}
	host := os.Getenv(TsdbHost)
	if host == "" {
		return nil, errors.New("No TSDB host found")
	}
	split := strings.Split(id, "/")
	if len(split) < 3 {
		return nil, errors.New("Bad APP_ID id")
	}
	c := new(config)
	c.AppName = split[len(split)-1]
	c.Group = split[1]
	c.Host = fmt.Sprintf("%s:4242", host)
	return c, nil
}

// Telescope documents here
type Telescope struct {
	Pipe       chan *client.DataPoint
	Conn       client.Client
	Config     *config
	CommonTags map[string]string
	Enabled    bool
}

func (self *Telescope) init() *Telescope {
	self.Pipe = make(chan *client.DataPoint, 10e6)
	var err error
	self.Config, err = envConfig()
	if err != nil {
		return self
	}
	self.Enabled = true
	self.CommonTags = make(map[string]string)
	self.CommonTags["app"] = self.Config.AppName
	self.CommonTags["group"] = self.Config.Group
	go self.loop()
	return self
}

func NewTelescope() *Telescope {
	return new(Telescope).init()
}

func (self *Telescope) connect() client.Client {
	opentsdbCfg := opentsdbConf.OpenTSDBConfig{
		OpentsdbHost: self.Config.Host,
	}
	self.Conn, _ = client.NewClient(opentsdbCfg)
	return self.Conn

}
func (self *Telescope) loop() {
	var batch []client.DataPoint
	deadline := time.After(Interval)
	for {
		select {
		case v := <-self.Pipe:
			batch = append(batch, *v)
		case <-deadline:
			deadline = time.After(Interval)
			conn := self.connect()
			if conn == nil {
				continue
			}
			if len(batch) == 0 {
				continue
			}
			fmt.Println("Sending metrics")
			for len(batch) > BatchSize {
				if resp, err := conn.Put(batch[:BatchSize], "details"); err != nil {
					fmt.Printf("Error occurs when putting datapoints: %v; %v\n", batch[:BatchSize], err)
				} else {
					fmt.Printf("Err  %s \n", resp.String())
				}
				batch = batch[BatchSize:]
			}
		}
	}
}
