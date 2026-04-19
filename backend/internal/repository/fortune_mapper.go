package repository

import (
	"context"
	"database/sql"
	"time"
)

type FortuneMapper interface {
	GetByUserAndDate(ctx context.Context, userID int64, date time.Time) (string, error)
	Save(ctx context.Context, userID int64, date time.Time, content string) error
	CleanupBefore(ctx context.Context, cutoffDate time.Time) error
}

type fortuneMapper struct {
	db *sql.DB
}

// NewFortuneMapper 创建运势相关数据访问对象。
// 参数：db - 数据库连接池。
// 返回：FortuneMapper - 运势 Mapper 接口实现。
func NewFortuneMapper(db *sql.DB) FortuneMapper {
	return &fortuneMapper{db: db}
}

// GetByUserAndDate 按用户和日期查询运势（占位实现）。
// 参数：ctx - 上下文；userID - 用户ID；date - 目标日期。
// 返回：string - 运势文本；error - 查询错误。
func (m *fortuneMapper) GetByUserAndDate(ctx context.Context, userID int64, date time.Time) (string, error) {
	_ = ctx
	_ = userID
	_ = date
	return "", sql.ErrNoRows
}

// Save 保存运势记录（占位实现）。
// 参数：ctx - 上下文；userID - 用户ID；date - 目标日期；content - 运势内容。
// 返回：error - 写入失败错误。
func (m *fortuneMapper) Save(ctx context.Context, userID int64, date time.Time, content string) error {
	_ = ctx
	_ = userID
	_ = date
	_ = content
	return nil
}

// CleanupBefore 清理指定日期之前的数据（占位实现）。
// 参数：ctx - 上下文；cutoffDate - 截止日期。
// 返回：error - 删除失败错误。
func (m *fortuneMapper) CleanupBefore(ctx context.Context, cutoffDate time.Time) error {
	_ = ctx
	_ = cutoffDate
	return nil
}
