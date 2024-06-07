package cache

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/mitchellh/mapstructure"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/server/options"
)

const typeRedis = "redis"

type redisClient struct {
	client *redis.Client
}

// redisOptions used to create a redis client.
type redisOptions struct {
	Host     string `json:"host" yaml:"host" mapstructure:"host"`
	Port     int    `json:"port" yaml:"port" mapstructure:"port"`
	Password string `json:"password" yaml:"password" mapstructure:"password"`
	DB       int    `json:"db" yaml:"db" mapstructure:"db"`
}

func NewRedisClient(option *redisOptions, stopCh <-chan struct{}) (Interface, error) {
	var r redisClient

	redisOptions := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", option.Host, option.Port),
		Password: option.Password,
		DB:       option.DB,
	}

	if stopCh == nil {
		klog.Fatalf("no stop channel passed, redis connections will leak.")
	}

	r.client = redis.NewClient(redisOptions)

	if err := r.client.Ping().Err(); err != nil {
		r.client.Close()
		return nil, err
	}

	// close redis in case of connection leak
	if stopCh != nil {
		go func() {
			<-stopCh
			if err := r.client.Close(); err != nil {
				klog.Error(err)
			}
		}()
	}

	return &r, nil
}

func (r *redisClient) Get(key string) (string, error) {
	return r.client.Get(key).Result()
}

func (r *redisClient) Keys(pattern string) ([]string, error) {
	return r.client.Keys(pattern).Result()
}

func (r *redisClient) Set(key string, value string, duration time.Duration) error {
	return r.client.Set(key, value, duration).Err()
}

func (r *redisClient) Del(keys ...string) error {
	return r.client.Del(keys...).Err()
}

func (r *redisClient) Exists(keys ...string) (bool, error) {
	existedKeys, err := r.client.Exists(keys...).Result()
	if err != nil {
		return false, err
	}

	return len(keys) == int(existedKeys), nil
}

func (r *redisClient) Expire(key string, duration time.Duration) error {
	return r.client.Expire(key, duration).Err()
}

type redisFactory struct{}

func (rf *redisFactory) Type() string {
	return typeRedis
}

func (rf *redisFactory) Create(options options.DynamicOptions, stopCh <-chan struct{}) (Interface, error) {
	var rOptions redisOptions
	if err := mapstructure.Decode(options, &rOptions); err != nil {
		return nil, err
	}
	if rOptions.Port == 0 {
		return nil, errors.New("invalid service port number")
	}
	if len(rOptions.Host) == 0 {
		return nil, errors.New("invalid service host")
	}
	client, err := NewRedisClient(&rOptions, stopCh)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func init() {
	RegisterCacheFactory(&redisFactory{})
}
