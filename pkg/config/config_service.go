package config

import (
	"context"
)

// ConfigService は UIStateStore の Wails バインディングラッパー。
// フロントエンドから UIStateStore へのアクセスを提供する。
type ConfigService struct {
	store UIStateStore
}

// NewConfigService は新しい ConfigService を生成する。
func NewConfigService(store UIStateStore) *ConfigService {
	return &ConfigService{store: store}
}

// UIStateGetJSON は指定した namespace/key の JSON 値を文字列として返す。
// フロントエンドで JSON.parse すること。
// キーが存在しない場合は空文字を返す。
func (s *ConfigService) UIStateGetJSON(namespace, key string) (string, error) {
	return s.store.Get(context.Background(), namespace, key)
}

// UIStateSetJSON は指定した namespace/key に JSON 値を保存する。
// フロントエンドからは任意のオブジェクトが渡されるので、ここで any として受け取る。
func (s *ConfigService) UIStateSetJSON(namespace, key string, value any) error {
	return s.store.SetJSON(context.Background(), namespace, key, value)
}

// UIStateDelete は指定した namespace/key のエントリを削除する。
func (s *ConfigService) UIStateDelete(namespace, key string) error {
	return s.store.Delete(context.Background(), namespace, key)
}
