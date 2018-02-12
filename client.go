package telescope

import (
	"strings"
	"time"

	"github.com/bluebreezecf/opentsdb-goclient/client"
	"github.com/gin-gonic/gin"
)

// Timer documents here
//  t := NewTimer("Latency")
//  defer t.Stop()
type Timer struct {
	startedAt time.Time
	name      string
	tags      map[string]string
}

func (self *Timer) init(name string, tags map[string]string) *Timer {
	self.startedAt = time.Now()
	self.name = name
	self.tags = tags
	return self
}

func NewTimer(name string, tags map[string]string) *Timer {
	return new(Timer).init(name, tags)
}

func (self *Timer) Stop() {
	duration := time.Since(self.startedAt)
	Put(self.name, float32(duration), self.tags)
	return
}

func Put(name string, value float32, tags map[string]string) {
	p := new(client.DataPoint)
	p.Metric = name
	p.Value = value
	p.Timestamp = time.Now().Unix()
	p.Tags = merge(telescope.CommonTags, tags)
	// p.Tags = telescope.CommonTags
	if !telescope.Enabled {
		return
	}
	telescope.Pipe <- p
	return
}

func merge(a map[string]string, b map[string]string) map[string]string {
	if b == nil {
		return a
	}
	for k, v := range a {
		b[k] = v
	}
	return b
}

func ScrabContext(c *gin.Context) map[string]string {
	tags := make(map[string]string)
	if c == nil {
		return tags
	}
	tags["Method"] = c.Request.Method
	// tags["RemoteAddr"] = c.Request.RemoteAddr
	tags["RequestURI"] = strings.Replace(c.Request.RequestURI, "$", "", -1)
	// tags["Host"] = c.Request.Host
	return tags
}
