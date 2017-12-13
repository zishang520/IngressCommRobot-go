package Config

import (
	"encoding/json"
	filelock "github.com/juju/go4/lock"
	"gopkg.in/retry.v1"
	"io"
	"os"
	"sync"
	"time"
)

type Config map[string]interface{}

type Options struct {
	Config
	filename string
	mu       sync.Mutex
}

var noOptions Options

func New(filename string) (*Options, error) {
	o := &Options{
		filename: filename,
		Config:   make(Config),
	}
	locked, err := lockFile(lockFileName(o.filename))
	if err != nil {
		return o, err
	}
	defer locked.Close()
	f, err := os.Open(o.filename)
	if err != nil {
		return o, err
	}
	defer f.Close()
	if o.mergeFrom(f) != nil {
		return o, err
	}
	return o, nil
}

func (o *Options) Set(key string, value interface{}) *Options {
	o.Config[key] = value
	return o
}

func (o *Options) Get(key string) interface{} {
	if value, ok := o.Config[key]; ok {
		return value
	}
	return nil
}

func (o *Options) Has(key string) bool {
	_, ok := o.Config[key]
	return ok
}

func (o *Options) Delete(key string) *Options {
	delete(o.Config, key)
	return o
}

func (o *Options) Clear() *Options {
	o.Config = make(Config)
	return o
}

func (o *Options) writeTo(w io.Writer) error {
	if err := json.NewEncoder(w).Encode(o.Config); err != nil {
		return err
	}
	return nil
}

func (o *Options) merge(config Config) {
	for k, v := range config {
		o.Config[k] = v
	}
}

func (o *Options) mergeFrom(r io.Reader) error {
	decoder := json.NewDecoder(r)
	var data json.RawMessage
	if err := decoder.Decode(&data); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil
	}
	o.merge(config)
	return nil
}
func (o *Options) Save() (*Options, error) {
	locked, err := lockFile(lockFileName(o.filename))
	if err != nil {
		return o, err
	}
	defer locked.Close()
	f, err := os.OpenFile(o.filename, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return o, err
	}
	defer f.Close()
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := f.Truncate(0); err != nil {
		return o, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return o, err
	}
	return o, o.writeTo(f)
}

func lockFileName(path string) string {
	return path + ".lock"
}

var attempt = retry.LimitTime(3*time.Second, retry.Exponential{
	Initial:  100 * time.Microsecond,
	Factor:   1.5,
	MaxDelay: 100 * time.Millisecond,
})

func lockFile(path string) (io.Closer, error) {
	for a := retry.Start(attempt, nil); a.Next(); {
		locker, err := filelock.Lock(path)
		if err == nil {
			return locker, nil
		}
		if !a.More() {
			return nil, err
		}
	}
	panic("unreachable")
}
