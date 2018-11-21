package opc

import (
	"io"
	"sync"
	"time"
)

//data holds the data structure that is refreshed with OPC data.
type data struct {
	tags map[string]interface{}
	mu   sync.RWMutex
}

//Get is the thread-safe getter for the tags.
func (d *data) Get(key string) (interface{}, bool) {
	d.mu.RLock()
	value, ok := d.tags[key]
	d.mu.RUnlock()
	return value, ok
}

//update is a helper function to update map
func (d *data) update(conn OpcConnection) {
	update := conn.Read()
	d.mu.Lock()
	for key, value := range update {
		d.tags[key] = value
	}
	d.mu.Unlock()
}

//Sync synchronizes the opc server and stores the data into the data model.
func (d *data) Sync(conn OpcConnection, refreshRate time.Duration) io.Closer {

        control := NewControl()
        ticker := time.NewTicker(refreshRate)
        
        d.update(conn)
	
        go func() {
		for {
			select {
			case <-ticker.C:
				d.update(conn)
			case <-control.close:
				ticker.Stop()
				control.done <- true
				return
			}
		}
	}()

	return control
}

//NewDataModel returns an OPC Data struct.
func NewDataModel() data {
	return data{tags: make(map[string]interface{})}
}

type control struct {
	close chan bool
	done  chan bool
}

func (c *control) Close() error {
	if c.close != nil && c.done != nil {
		c.close <- true
		<-c.done
	}
	return nil
}

func NewControl() *control {
        return &control{close: make(chan bool), done: make(chan bool)}
}