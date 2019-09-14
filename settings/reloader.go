package settings

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	consul "github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

const (
	consulWaitTime     = 60 * time.Second
	consulFallbackTime = time.Second
)

// Reloader provides current application configuration for other packages.
//
// It reloads settings from Consul.
type Reloader struct {
	c *consul.Client
	l *zap.Logger

	ready chan struct{}
	once  sync.Once

	rw        sync.RWMutex
	current   *Settings
	consulKey string
}

func ConnectAndRunReloader(ctx context.Context, consulAddrs, consulKey string) (*Reloader, *consul.Client) {
	consulClient, err := consul.NewClient(&consul.Config{
		Address: consulAddrs,
	})
	if err != nil {
		zap.L().Panic("Failed to create Consul client.", zap.Error(err))
	}
	reloader := NewReloader(consulClient, consulKey)
	go reloader.Run(ctx)
	zap.L().Info("Waiting for settings from Consul...")
	_ = reloader.Settings()
	zap.L().Info("Consul - Connected and got settings!")

	return reloader, consulClient
}

func NewReloader(consulClient *consul.Client, consulKey string) *Reloader {
	return &Reloader{
		c:         consulClient,
		l:         zap.L().Named("reloader"),
		ready:     make(chan struct{}),
		consulKey: consulKey,
	}
}

func (r *Reloader) Run(ctx context.Context) {
	kv := r.c.KV()
	opts := &consul.QueryOptions{
		WaitTime: consulWaitTime,
	}
	opts = opts.WithContext(ctx)

	for ctx.Err() == nil {
		// wait for settings until they are changed up to consulWaitTime
		p, _, err := kv.Get(r.consulKey, opts)
		if err != nil {
			r.l.Warn("Failed to get settings.", zap.Error(err))
			time.Sleep(consulFallbackTime)
			continue
		}
		if p == nil {
			r.l.Warn("Failed to get settings - key not found.", zap.String("key", r.consulKey))
			time.Sleep(consulFallbackTime)
			continue
		}

		// check if settings were changed
		switch {
		case opts.WaitIndex == p.ModifyIndex:
			r.l.Debug("Settings not changed.", zap.Uint64("ModifyIndex", p.ModifyIndex), zap.Uint64("WaitIndex", opts.WaitIndex))
			continue
		case opts.WaitIndex > p.ModifyIndex:
			r.l.DPanic("Stale settings!", zap.Uint64("ModifyIndex", p.ModifyIndex), zap.Uint64("WaitIndex", opts.WaitIndex))
		}

		// unmarshal settings
		var new Settings
		if err = json.Unmarshal(p.Value, &new); err != nil {
			r.l.Warn("Failed to unmarshal settings.", zap.Error(err), zap.ByteString("settings", p.Value))
			time.Sleep(consulFallbackTime)
			continue
		}

		// remember current index, apply new settings, signal that we have them
		r.l.Info("Applying new settings.", zap.Uint64("ModifyIndex", p.ModifyIndex), zap.Reflect("settings", new))
		opts.WaitIndex = p.ModifyIndex
		r.rw.Lock()
		r.current = &new
		r.rw.Unlock()
		r.once.Do(func() {
			close(r.ready)
		})
	}

	r.l.Info("Exiting.", zap.Error(ctx.Err()))
}

// Settings returns current settings.
//
// If no settings are available yet, this method blocks.
//
// Caller should not modify returned settings in any way.
func (r *Reloader) Settings() *Settings {
	<-r.ready

	r.rw.RLock()
	s := r.current
	r.rw.RUnlock()
	return s
}

// PutSettings stores given settings in Consul.
func (r *Reloader) PutSettings(s *Settings) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	kv := &consul.KVPair{
		Key:   r.consulKey,
		Value: b,
	}
	_, err = r.c.KV().Put(kv, nil)
	return err
}
