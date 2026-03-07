package config

import (
	"context"
)

// ConfigService は UIStateStore の Wails バインディングラッパー。
// フロントエンドから UIStateStore へのアクセスを提供する。
type ConfigService struct {
	uiStateStore UIStateStore
	configStore  Config
}

// NewConfigService は新しい ConfigService を生成する。
func NewConfigService(store *SQLiteStore) *ConfigService {
	return &ConfigService{
		uiStateStore: store,
		configStore:  store,
	}
}

// UIStateGetJSON は指定した namespace/key の JSON 値を文字列として返す。
// フロントエンドで JSON.parse すること。
// キーが存在しない場合は空文字を返す。
func (s *ConfigService) UIStateGetJSON(namespace, key string) (string, error) {
	return s.uiStateStore.Get(context.Background(), namespace, key)
}

// UIStateSetJSON は指定した namespace/key に JSON 値を保存する。
// フロントエンドからは任意のオブジェクトが渡されるので、ここで any として受け取る。
func (s *ConfigService) UIStateSetJSON(namespace, key string, value any) error {
	return s.uiStateStore.SetJSON(context.Background(), namespace, key, value)
}

// UIStateDelete は指定した namespace/key のエントリを削除する。
func (s *ConfigService) UIStateDelete(namespace, key string) error {
	return s.uiStateStore.Delete(context.Background(), namespace, key)
}

// ConfigGet は指定 namespace/key の設定値を取得する。
func (s *ConfigService) ConfigGet(namespace, key string) (string, error) {
	return s.configStore.Get(context.Background(), namespace, key)
}

// ConfigSet は指定 namespace/key へ値を保存する。
func (s *ConfigService) ConfigSet(namespace, key, value string) error {
	return s.configStore.Set(context.Background(), namespace, key, value)
}

// ConfigSetMany は指定 namespace に複数キーをまとめて保存する。
func (s *ConfigService) ConfigSetMany(namespace string, values map[string]string) error {
	ctx := context.Background()
	for k, v := range values {
		if err := s.configStore.Set(ctx, namespace, k, v); err != nil {
			return err
		}
	}
	return nil
}

// ConfigDelete は指定 namespace/key を削除する。
func (s *ConfigService) ConfigDelete(namespace, key string) error {
	return s.configStore.Delete(context.Background(), namespace, key)
}

// ConfigGetAll は指定 namespace の全キーを取得する。
func (s *ConfigService) ConfigGetAll(namespace string) (map[string]string, error) {
	return s.configStore.GetAll(context.Background(), namespace)
}
