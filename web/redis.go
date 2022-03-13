package web

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	Client *redis.Client
}

func (r *Redis) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	err := r.Client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}

func (r *Redis) Incr(ctx context.Context, key string) error {
	err := r.Client.Incr(ctx, key).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) Exists(ctx context.Context, keys ...string) (bool, error) {
	num, err := r.Client.Exists(ctx, keys...).Result()
	if err != nil {
		return false, err
	}

	return num > 0, nil
}

func (r *Redis) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := r.Client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (r *Redis) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	values, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	return values, nil
}
