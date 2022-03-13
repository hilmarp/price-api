package scraper

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"bitbucket.org/hilmarp/price-scraper/metrics"
	"github.com/go-redis/redis/v8"
)

// Redis implements the redis storage backend for Colly
type Redis struct {
	// RClient has the redis client and app context
	Client *redis.Client

	// Prefix is an optional string in the keys. It can be used
	// to use one redis database for independent scraping tasks.
	Prefix string

	// Expiration time for Visited keys. After expiration pages
	// are to be visited again.
	Expires time.Duration

	// Only used for cookie methods.
	mu sync.RWMutex
}

// Init initializes the redis storage
func (s *Redis) Init() error {
	if s.Client == nil {
		return fmt.Errorf("Redis storage client not set")
	}

	return nil
}

// Visited implements colly/storage.Visited()
func (s *Redis) Visited(requestID uint64) error {
	return s.Client.Set(context.TODO(), s.getIDStr(requestID), "1", s.Expires).Err()
}

// IsVisited implements colly/storage.IsVisited()
func (s *Redis) IsVisited(requestID uint64) (bool, error) {
	_, err := s.Client.Get(context.TODO(), s.getIDStr(requestID)).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// Cookies implements colly/storage.Cookies()
func (s *Redis) Cookies(u *url.URL) string {
	// TODO(js) Cookie methods currently have no way to return an error.

	s.mu.RLock()
	cookiesStr, err := s.Client.Get(context.TODO(), s.getCookieID(u.Host)).Result()
	s.mu.RUnlock()
	if err == redis.Nil {
		cookiesStr = ""
	} else if err != nil {
		// return nil, err
		log.Printf("Cookies() .Get error %s", err)
		return ""
	}
	return cookiesStr
}

// SetCookies implements colly/storage..SetCookies()
func (s *Redis) SetCookies(u *url.URL, cookies string) {
	// TODO(js) Cookie methods currently have no way to return an error.

	// We need to use a write lock to prevent a race in the db:
	// if two callers set cookies in a very small window of time,
	// it is possible to drop the new cookies from one caller
	// ('last update wins' == best avoided).
	s.mu.Lock()
	defer s.mu.Unlock()
	// return s.Client.Set(s.getCookieID(u.Host), stringify(cnew), 0).Err()
	err := s.Client.Set(context.TODO(), s.getCookieID(u.Host), cookies, 0).Err()
	if err != nil {
		// return nil
		log.Printf("SetCookies() .Set error %s", err)
		return
	}
}

// Clear removes all entries from the storage
func (s *Redis) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Client.FlushDB(context.TODO()).Err()
}

// ClearByPrefix removes all entries with Prefix from the storage
func (s *Redis) ClearByPrefix() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.Client.Keys(context.TODO(), s.getCookieID("*"))
	keys, err := r.Result()
	if err != nil {
		return err
	}
	r2 := s.Client.Keys(context.TODO(), s.Prefix+":request:*")
	keys2, err := r2.Result()
	if err != nil {
		return err
	}
	keys = append(keys, keys2...)
	keys = append(keys, s.getQueueID())
	return s.Client.Del(context.TODO(), keys...).Err()
}

// AddRequest implements queue.Storage.AddRequest() function
func (s *Redis) AddRequest(r []byte) error {
	err := s.Client.RPush(context.TODO(), s.getQueueID(), r).Err()
	if err != nil {
		metrics.ScraperRedisEnqueuesError.Inc()
		return err
	}
	metrics.ScraperRedisEnqueues.Inc()
	return nil
}

// GetRequest implements queue.Storage.GetRequest() function
func (s *Redis) GetRequest() ([]byte, error) {
	r, err := s.Client.LPop(context.TODO(), s.getQueueID()).Bytes()
	if err != nil {
		metrics.ScraperRedisDequeuesError.Inc()
		return nil, err
	}
	metrics.ScraperRedisDequeues.Inc()
	return r, err
}

// QueueSize implements queue.Storage.QueueSize() function
func (s *Redis) QueueSize() (int, error) {
	i, err := s.Client.LLen(context.TODO(), s.getQueueID()).Result()
	return int(i), err
}

func (s *Redis) getIDStr(ID uint64) string {
	return fmt.Sprintf("%s:request:%d", s.Prefix, ID)
}

func (s *Redis) getCookieID(c string) string {
	return fmt.Sprintf("%s:cookie:%s", s.Prefix, c)
}

func (s *Redis) getQueueID() string {
	return fmt.Sprintf("%s:queue", s.Prefix)
}
