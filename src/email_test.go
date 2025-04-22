package main

import (
	"database/sql"    // 导入数据库 SQL 包
	"encoding/json" // 导入 JSON 编码/解码包
	"testing"         // 导入 Go 的测试包
	"time"            // 导入时间包

	"github.com/stretchr/testify/assert" // 导入 testify 断言库
)

// insertUserEmailVerificationRequest 是一个测试辅助函数，用于向数据库中插入一条用户邮箱验证请求记录。
// 注意：此函数的 SQL 语句参数似乎存在问题 (占位符数量与提供的值不匹配)。
// 参数：
//   db (*sql.DB): 数据库连接对象。
//   request (*UserEmailVerificationRequest): 要插入的验证请求数据。
// 返回值：
//   error: 如果数据库操作出错，则返回错误信息，否则返回 nil。
func insertUserEmailVerificationRequest(db *sql.DB, request *UserEmailVerificationRequest) error {
	// SQL 语句的 VALUES 子句有 7 个 '?' 占位符，但只提供了 6 个参数。
	// 最后三个参数 request.CreatedAt.Unix(), request.Code, request.UserId 看起来是多余或错误的。
	// 正确的语句可能只需要 request.UserId, request.CreatedAt.Unix(), request.ExpiresAt.Unix(), request.Code。
	_, err := db.Exec("INSERT INTO user_email_verification_request (user_id, created_at, expires_at, code) VALUES (?, ?, ?, ?)", request.UserId, request.CreatedAt.Unix(), request.ExpiresAt.Unix(), request.Code) // 修正后的参数列表推测
	// 原始问题代码: _, err := db.Exec("INSERT INTO user_email_verification_request (user_id, created_at, expires_at, code) VALUES (?, ?, ?, ?)", request.UserId, request.CreatedAt.Unix(), request.ExpiresAt.Unix(), request.Code, request.CreatedAt.Unix(), request.Code, request.UserId)
	return err
}

// TestEncodeEmailToJSON 测试 encodeEmailToJSON 函数是否能正确地将邮箱地址编码为 JSON 格式。
// 它创建一个简单的邮箱字符串，调用 encodeEmailToJSON，然后将返回的 JSON 字符串解码回 EmailJSON 结构体，
// 最后断言解码后的结构体与预期的结构体相等。
func TestEncodeEmailToJSON(t *testing.T) {
	t.Parallel() // 标记此测试可以与其他 Parallel 测试并行运行

	email := "user@example.com" // 测试用的邮箱地址

	// 预期得到的 EmailJSON 结构体
	expected := EmailJSON{
		Email: email,
	}

	var result EmailJSON // 用于存储解码后的结果

	// 调用被测试函数获取 JSON 字符串，然后解码到 result 结构体中
	err := json.Unmarshal([]byte(encodeEmailToJSON(email)), &result)
	assert.NoError(t, err) // 断言解码过程没有错误

	// 断言解码后的结果与预期结果相等
	assert.Equal(t, expected, result)
}

// TestEmailUpdateRequestEncodeToJSON 测试 EmailUpdateRequest 结构体的 EncodeToJSON 方法。
// 它创建一个 EmailUpdateRequest 实例，设置其字段值，然后调用 EncodeToJSON 方法。
// 接着，它将返回的 JSON 字符串解码回 EmailUpdateRequestJSON 结构体，
// 并断言解码后的结构体字段与原始请求对象的相应字段（特别是时间戳转换为 Unix 秒数后）相等。
func TestEmailUpdateRequestEncodeToJSON(t *testing.T) {
	t.Parallel() // 标记此测试可以与其他 Parallel 测试并行运行

	// 获取当前时间并截断纳秒，用于测试
	now := time.Unix(time.Now().Unix(), 0)

	// 创建一个测试用的 EmailUpdateRequest 实例
	request := EmailUpdateRequest{
		Id:        "1",
		UserId:    "1",
		Email:     "user@example.com",
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute), // 过期时间设为 10 分钟后
		Code:      "12345678",
	}

	// 预期得到的 EmailUpdateRequestJSON 结构体，时间字段应为 Unix 时间戳
	expected := EmailUpdateRequestJSON{
		Id:            request.Id,
		UserId:        request.UserId,
		Email:         request.Email,
		CreatedAtUnix: request.CreatedAt.Unix(),
		ExpiresAtUnix: request.ExpiresAt.Unix(),
		Code:          request.Code,
	}

	var result EmailUpdateRequestJSON // 用于存储解码后的结果

	// 调用被测试对象的 EncodeToJSON 方法获取 JSON 字符串，然后解码到 result 结构体中
	err := json.Unmarshal([]byte(request.EncodeToJSON()), &result)
	assert.NoError(t, err) // 断言解码过程没有错误

	// 断言解码后的结果与预期结果相等
	assert.Equal(t, expected, result)
}

// TestUserEmailVerificationRequestEncodeToJSON 测试 UserEmailVerificationRequest 结构体的 EncodeToJSON 方法。
// 这个测试与 TestEmailUpdateRequestEncodeToJSON 类似，但针对的是 UserEmailVerificationRequest 类型。
// 它创建实例，调用 EncodeToJSON，解码返回的 JSON，并断言结果的正确性。
func TestUserEmailVerificationRequestEncodeToJSON(t *testing.T) {
	t.Parallel() // 标记此测试可以与其他 Parallel 测试并行运行

	// 获取当前时间并截断纳秒，用于测试
	now := time.Unix(time.Now().Unix(), 0)

	// 创建一个测试用的 UserEmailVerificationRequest 实例
	request := UserEmailVerificationRequest{
		UserId:    "1",
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute), // 过期时间设为 10 分钟后
		Code:      "12345678",
	}

	// 预期得到的 UserEmailVerificationRequestJSON 结构体，时间字段应为 Unix 时间戳
	expected := UserEmailVerificationRequestJSON{
		UserId:        request.UserId,
		CreatedAtUnix: request.CreatedAt.Unix(),
		ExpiresAtUnix: request.ExpiresAt.Unix(),
		Code:          request.Code,
	}

	var result UserEmailVerificationRequestJSON // 用于存储解码后的结果

	// 调用被测试对象的 EncodeToJSON 方法获取 JSON 字符串，然后解码到 result 结构体中
	err := json.Unmarshal([]byte(request.EncodeToJSON()), &result)
	assert.NoError(t, err) // 断言解码过程没有错误

	// 断言解码后的结果与预期结果相等
	assert.Equal(t, expected, result)
}

// EmailJSON 是用于在测试中表示只包含 email 字段的 JSON 结构。
type EmailJSON struct {
	Email string `json:"email"` // 邮箱地址，对应 JSON 中的 "email" 键
}

// EmailUpdateRequestJSON 是用于在测试中表示 EmailUpdateRequest 编码为 JSON 后的结构。
// 注意时间字段是以 Unix 时间戳 (int64) 的形式表示的。
type EmailUpdateRequestJSON struct {
	Id            string `json:"id"`         // 请求 ID，对应 JSON 中的 "id" 键
	UserId        string `json:"user_id"`    // 用户 ID，对应 JSON 中的 "user_id" 键
	Email         string `json:"email"`      // 邮箱地址，对应 JSON 中的 "email" 键
	CreatedAtUnix int64  `json:"created_at"` // 创建时间的 Unix 时间戳，对应 JSON 中的 "created_at" 键
	ExpiresAtUnix int64  `json:"expires_at"` // 过期时间的 Unix 时间戳，对应 JSON 中的 "expires_at" 键
	Code          string `json:"code"`       // 验证码，对应 JSON 中的 "code" 键
}

// UserEmailVerificationRequestJSON 是用于在测试中表示 UserEmailVerificationRequest 编码为 JSON 后的结构。
// 同样，时间字段是以 Unix 时间戳 (int64) 的形式表示的。
type UserEmailVerificationRequestJSON struct {
	UserId        string `json:"user_id"`    // 用户 ID，对应 JSON 中的 "user_id" 键
	CreatedAtUnix int64  `json:"created_at"` // 创建时间的 Unix 时间戳，对应 JSON 中的 "created_at" 键
	ExpiresAtUnix int64  `json:"expires_at"` // 过期时间的 Unix 时间戳，对应 JSON 中的 "expires_at" 键
	Code          string `json:"code"`       // 验证码，对应 JSON 中的 "code" 键
}
