package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/UmedjonQurbonov/cinema-user-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

// DTO для сериализации в JSON внутри Redis
type PendingUserPayload struct {
	Code string       `json:"code"`
	User *domain.User `json:"user"`
}

type TempUserRepository struct {
	client *redis.Client
}

func NewTempUserRepository(client *redis.Client) *TempUserRepository {
	return &TempUserRepository{
		client: client,
	}
}

// 1. Сохранить пользователя и код по ключу "pending_user:" + email с заданным TTL
func (r *TempUserRepository) SavePendingUser(ctx context.Context, email, code string, user *domain.User, ttl time.Duration) error {
	// 1. Собираем структуру
	userPayload := PendingUserPayload{
		Code: code,
		User: user,
	}

	// 2. Сериализуем в JSON
	data, err := json.Marshal(userPayload)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	} 

	// 3. Формируем уникальный ключ
	key := fmt.Sprintf("pending_user:%s", email)

	// 4. Записываем в Redis: (ctx, key, value, ttl) и проверяем .Err()
	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis.Set: %w", err)
	}

	return nil
}

// 2. Достать JSON по ключу, распарсить и вернуть (*domain.User, code, error)
func (r *TempUserRepository) GetPendingUser(ctx context.Context, email string) (*domain.User, string, error) {
	// ТВОЯ ЛОГИКА:
	// 1. Сделать r.client.Get(ctx, key).Result()
	key := fmt.Sprintf("pending_user:%s", email)
	result, err := r.client.Get(ctx, key).Result()
	// 2. Если ошибка (например, ключ протух или не существует) — вернуть ошибку
	if err != nil {
		return &domain.User{}, "", fmt.Errorf("r.client.Get: %w", err)
	}
	// 3. Распарсить JSON в pendingUserPayload через json.Unmarshal
	var payload PendingUserPayload
	if err := json.Unmarshal([]byte(result), &payload); err != nil {
		return nil, "", fmt.Errorf("json.Unmarshal: %w", err)
	}
	// 4. Вернуть payload.User, payload.Code, nil
	return payload.User, payload.Code, nil
}

// 3. Удалить ключ по email
func (r *TempUserRepository) DeletePendingUser(ctx context.Context, email string) error {
	// ТВОЯ ЛОГИКА:
	key := fmt.Sprintf("pending_user:%s",email)
	// Вызвать r.client.Del(ctx, key).Err()
	err := r.client.Del(ctx, key).Err()

	if err != nil {
		return fmt.Errorf("r.client.Del: %w", err)
	}
	
	return nil
}