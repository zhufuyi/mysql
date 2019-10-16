package mysql

import (
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	db     *gorm.DB
	tables []interface{} // 各个对象指针地址集合

	// ErrNotFound 空记录
	ErrNotFound = errors.New("record not found")
)

// Init 初始化mysql
func Init(addr string, isEnableLog ...bool) error {
	var err error

	db, err = gorm.Open("mysql", addr)
	if err != nil {
		return err
	}

	db.DB().SetMaxIdleConns(3)                  // 空闲连接数
	db.DB().SetMaxOpenConns(100)                // 最大连接数
	db.DB().SetConnMaxLifetime(3 * time.Minute) // 3分钟后断开多余的空闲连接
	db.SingularTable(true)                      // 保持表名和对象名一致

	if len(isEnableLog) == 1 && isEnableLog[0] {
		db.LogMode(true)              // 开启日志
		db.SetLogger(newGormLogger()) // 自定义日志
	}

	if err = db.DB().Ping(); err != nil {
		return err
	}

	SyncTable()

	return nil
}

// GetDB 获取连接
func GetDB() *gorm.DB {
	if db == nil {
		panic("db is nil, please reconnect mysql.")
	}
	return db
}

// AddTables 添加表
func AddTables(object ...interface{}) {
	tables = append(tables, object...)
}

// SyncTable 同步表
func SyncTable() {
	GetDB().AutoMigrate(tables...) // 确保对象和mysql表一致，只支持自动添加新的列，对于存在的列不可以修改列属性
}

// Model 表内嵌字段
type Model struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `sql:"index" json:"deletedAt"`
}

// KV map类型
type KV map[string]interface{}

// TxRecover 回收事务过程中的panic，使用时在前面添加defer关键字，例如：defer TxRecover(tx)
func TxRecover(tx *gorm.DB) {
	if r := recover(); r != nil {
		fmt.Printf("transaction failed, err = %v\n", r)
		tx.Rollback()
	}
}
