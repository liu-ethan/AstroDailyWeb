package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type UserProfile struct {
	UserID        int64  `json:"user_id"`
	IsSubscribed  bool   `json:"is_subscribed"`
	Birthday      string `json:"birthday"`
	Constellation string `json:"constellation"`
	Gender        string `json:"gender"`
	City          string `json:"city"`
	Occupation    string `json:"occupation"`
}

type UserMapper interface {
	UpdateSubscription(ctx context.Context, userID int64, subscribed bool) error
	ListSubscribedUsers(ctx context.Context) ([]UserRecord, error)
	GetProfile(ctx context.Context, userID int64) (UserProfile, error)
	UpsertProfile(ctx context.Context, profile UserProfile) error
}

type userMapper struct {
	db *sql.DB
}

// NewUserMapper 创建用户设置相关数据访问对象。
// 参数：db - 数据库连接池。
// 返回：UserMapper - 用户 Mapper 接口实现。
func NewUserMapper(db *sql.DB) UserMapper {
	return &userMapper{db: db}
}

// UpdateSubscription 更新用户订阅状态。
// 参数：ctx - 上下文；userID - 用户ID；subscribed - 是否订阅。
// 返回：error - 更新失败错误。
func (m *userMapper) UpdateSubscription(ctx context.Context, userID int64, subscribed bool) error {
	const stmt = `UPDATE users SET is_subscribed = ? WHERE id = ?`
	_, err := m.db.ExecContext(ctx, stmt, subscribed, userID)
	return err
}

// ListSubscribedUsers 查询已订阅用户列表。
// 参数：ctx - 上下文。
// 返回：[]UserRecord - 用户列表；error - 查询失败错误。
func (m *userMapper) ListSubscribedUsers(ctx context.Context) ([]UserRecord, error) {
	const query = `SELECT id, email, password, is_subscribed FROM users WHERE is_subscribed = 1`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	users := make([]UserRecord, 0)
	for rows.Next() {
		var u UserRecord
		if scanErr := rows.Scan(&u.ID, &u.Email, &u.Password, &u.IsSubscribed); scanErr != nil {
			return nil, scanErr
		}
		users = append(users, u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// GetProfile 查询用户资料。
// 参数：ctx - 上下文；userID - 用户ID。
// 返回：UserProfile - 用户资料；error - 查询失败错误。
func (m *userMapper) GetProfile(ctx context.Context, userID int64) (UserProfile, error) {
	const query = `
SELECT u.id, u.is_subscribed, up.birthday, up.constellation, up.gender, up.city, up.occupation
FROM users u
LEFT JOIN user_profiles up ON u.id = up.user_id
WHERE u.id = ?
LIMIT 1`
	var profile UserProfile
	var birthday sql.NullTime
	var constellation sql.NullString
	var gender sql.NullString
	var city sql.NullString
	var occupation sql.NullString
	err := m.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.UserID,
		&profile.IsSubscribed,
		&birthday,
		&constellation,
		&gender,
		&city,
		&occupation,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserProfile{}, nil
		}
		return UserProfile{}, err
	}
	if birthday.Valid {
		profile.Birthday = birthday.Time.Format("2006-01-02")
	}
	if constellation.Valid {
		profile.Constellation = constellation.String
	}
	if gender.Valid {
		profile.Gender = gender.String
	}
	if city.Valid {
		profile.City = city.String
	}
	if occupation.Valid {
		profile.Occupation = occupation.String
	}
	return profile, nil
}

// UpsertProfile 新增或更新用户资料。
// 参数：ctx - 上下文；profile - 用户资料。
// 返回：error - 写入失败错误。
func (m *userMapper) UpsertProfile(ctx context.Context, profile UserProfile) error {
	birthday, err := time.Parse("2006-01-02", profile.Birthday)
	if err != nil {
		return err
	}
	const stmt = `
INSERT INTO user_profiles (user_id, birthday, constellation, gender, city, occupation)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
birthday = VALUES(birthday),
constellation = VALUES(constellation),
gender = VALUES(gender),
city = VALUES(city),
occupation = VALUES(occupation)`
	_, err = m.db.ExecContext(ctx, stmt,
		profile.UserID,
		birthday,
		profile.Constellation,
		profile.Gender,
		profile.City,
		profile.Occupation,
	)
	return err
}
