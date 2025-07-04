package stores

import (
	"excalidraw-complete/core"
	"excalidraw-complete/stores/aws"
	"excalidraw-complete/stores/filesystem"
	"excalidraw-complete/stores/memory"
	"excalidraw-complete/stores/sqlite"
	"os"

	"github.com/sirupsen/logrus"
)

// Store 接口，包含所有存储组件
type Store interface {
	GetDocumentStore() core.DocumentStore
	GetUserStore() core.UserStore
	GetCanvasStore() core.CanvasStore
	GetSessionStore() core.SessionStore
}

// 基础存储结构
type BaseStore struct {
	Documents core.DocumentStore
	Users     core.UserStore
	Canvas    core.CanvasStore
	Sessions  core.SessionStore
}

func (s *BaseStore) GetDocumentStore() core.DocumentStore {
	return s.Documents
}

func (s *BaseStore) GetUserStore() core.UserStore {
	return s.Users
}

func (s *BaseStore) GetCanvasStore() core.CanvasStore {
	return s.Canvas
}

func (s *BaseStore) GetSessionStore() core.SessionStore {
	return s.Sessions
}

func GetStore() Store {
	storageType := os.Getenv("STORAGE_TYPE")
	
	storageField := logrus.Fields{
		"storageType": storageType,
	}

	switch storageType {
	case "filesystem":
		basePath := os.Getenv("LOCAL_STORAGE_PATH")
		storageField["basePath"] = basePath
		documentStore := filesystem.NewDocumentStore(basePath)
		// 对于文件系统存储，其他功能使用内存存储
		return &BaseStore{
			Documents: documentStore,
			Users:     memory.NewUserStore(),
			Canvas:    memory.NewCanvasStore(),
			Sessions:  memory.NewSessionStore(),
		}
	case "sqlite":
		dataSourceName := os.Getenv("DATA_SOURCE_NAME")
		storageField["dataSourceName"] = dataSourceName
		sqliteStore := sqlite.NewStore(dataSourceName)
		return &BaseStore{
			Documents: sqliteStore.Documents,
			Users:     sqliteStore.Users,
			Canvas:    sqliteStore.Canvas,
			Sessions:  sqliteStore.Sessions,
		}
	case "s3":
		bucketName := os.Getenv("S3_BUCKET_NAME")
		storageField["bucketName"] = bucketName
		documentStore := aws.NewDocumentStore(bucketName)
		// 对于S3存储，其他功能使用内存存储
		return &BaseStore{
			Documents: documentStore,
			Users:     memory.NewUserStore(),
			Canvas:    memory.NewCanvasStore(),
			Sessions:  memory.NewSessionStore(),
		}
	default:
		storageField["storageType"] = "in-memory"
		return &BaseStore{
			Documents: memory.NewDocumentStore(),
			Users:     memory.NewUserStore(),
			Canvas:    memory.NewCanvasStore(),
			Sessions:  memory.NewSessionStore(),
		}
	}
	
	logrus.WithFields(storageField).Info("Use storage")
}

// 为了向后兼容，保留GetDocumentStore函数
func GetDocumentStore() core.DocumentStore {
	return GetStore().GetDocumentStore()
}
