package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/CyberGeo335/pz9-redis-cache/internal/cache"
	"github.com/CyberGeo335/pz9-redis-cache/internal/config"
	"github.com/CyberGeo335/pz9-redis-cache/internal/task"
	"github.com/redis/go-redis/v9"
)

type TaskService struct {
	repo  *task.Repo
	redis *redis.Client
	cfg   config.Config
}

func NewTaskService(repo *task.Repo, redisClient *redis.Client, cfg config.Config) *TaskService {
	return &TaskService{
		repo:  repo,
		redis: redisClient,
		cfg:   cfg,
	}
}

func (s *TaskService) ListTasks() []task.Task {
	return s.repo.List()
}

func (s *TaskService) GetTaskByID(ctx context.Context, id int64) (task.Task, error) {
	key := cache.TaskByIDKey(id)

	if s.redis != nil {
		cached, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			var t task.Task
			if err := json.Unmarshal([]byte(cached), &t); err == nil {
				log.Println("cache hit:", key)
				return t, nil
			}

			log.Println("cache decode error, fallback to repo:", err)
		} else if errors.Is(err, redis.Nil) {
			log.Println("cache miss:", key)
		} else {
			log.Println("redis read error, fallback to repo:", err)
		}
	}

	t, err := s.repo.GetByID(id)
	if err != nil {
		return task.Task{}, err
	}

	if s.redis != nil {
		payload, err := json.Marshal(t)
		if err != nil {
			log.Println("cache encode error:", err)
			return t, nil
		}

		ttl := cache.TTLWithJitter(s.cfg.CacheTTL, s.cfg.CacheTTLJitter)
		if err := s.redis.Set(ctx, key, payload, ttl).Err(); err != nil {
			log.Println("redis write error:", err)
		} else {
			log.Println("cache set:", key, "ttl:", ttl)
		}
	}

	return t, nil
}

func (s *TaskService) CreateTask(ctx context.Context, t task.Task) task.Task {
	created := s.repo.Create(t)
	s.invalidateList(ctx)
	return created
}

func (s *TaskService) UpdateTask(ctx context.Context, t task.Task) error {
	if err := s.repo.Update(t); err != nil {
		return err
	}

	s.invalidateTask(ctx, t.ID)
	s.invalidateList(ctx)
	return nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id int64) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	s.invalidateTask(ctx, id)
	s.invalidateList(ctx)
	return nil
}

func (s *TaskService) invalidateTask(ctx context.Context, id int64) {
	if s.redis == nil {
		return
	}

	key := cache.TaskByIDKey(id)
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		log.Println("redis delete error:", err)
		return
	}

	log.Println("cache invalidated:", key)
}

func (s *TaskService) invalidateList(ctx context.Context) {
	if s.redis == nil {
		return
	}

	key := cache.TasksListKey()
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		log.Println("redis list delete error:", err)
		return
	}

	log.Println("cache invalidated:", key)
}
