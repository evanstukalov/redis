package store

import "time"

type ExpiredCollector struct {
	Store  *Store
	Ticker *time.Ticker
}

func NewExpiredCollector(store *Store) *ExpiredCollector {
	return &ExpiredCollector{
		Store:  store,
		Ticker: time.NewTicker(1 * time.Millisecond),
	}
}

func (expiredC *ExpiredCollector) Collect() {
	for key, value := range expiredC.Store.store {
		if value.ExpiredAt != nil && value.ExpiredAt.Before(time.Now()) {
			expiredC.Store.Remove(key)
		}
	}
}

func (expiredC *ExpiredCollector) Stop() {
	expiredC.Ticker.Stop()
}
