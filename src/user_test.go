package main

import (
	"encoding/json" // 导入 JSON 编码/解码包
	"testing"         // 导入 Go 的测试包
	"time"            // 导入时间包

	"github.com/stretchr/testify/assert" // 导入 testify 断言库
)

// TestUserEncodeToJSON 测试 User 结构体的 EncodeToJSON 方法。
// 这个测试旨在验证当 User 对象被编码为 JSON 时，只有特定的字段被包含，
// 特别是敏感信息如 PasswordHash 被排除在外，而时间戳被正确转换。
//
// 测试步骤：
// 1. 创建一个 User 实例，包含 ID, 创建时间, 密码哈希, 恢复码, 和 TOTP 注册状态。
// 2. 定义预期的 JSON 输出结构 (UserJSON)，它应包含 ID, 创建时间的 Unix 时间戳,
//    TOTP 注册状态, 以及恢复码，但不应包含 PasswordHash。
// 3. 调用 user.EncodeToJSON() 获取 JSON 字符串。
// 4. 将 JSON 字符串解码回 UserJSON 结构体。
// 5. 断言解码后的结构体与预期结构体完全相等，确保了正确的字段选择和格式转换。
func TestUserEncodeToJSON(t *testing.T) {
	t.Parallel() // 允许与其他 Parallel 测试并行运行

	// 获取当前时间并截断纳秒，用于测试时间戳
	now := time.Unix(time.Now().Unix(), 0)

	// 创建一个测试用的 User 实例
	user := User{
		Id:             "1",                           // 用户 ID
		CreatedAt:      now,                           // 创建时间
		PasswordHash:   "HASH1",                       // 密码哈希 (预期不包含在 JSON 中)
		RecoveryCode:   "12345678",                    // 恢复码 (预期包含在 JSON 中)
		TOTPRegistered: false,                         // TOTP 注册状态 (预期包含在 JSON 中)
	}

	// 预期得到的 JSON 结构，不包含 PasswordHash
	expected := UserJSON{
		Id:             user.Id,                       // 预期 ID 保持不变
		CreatedAtUnix:  user.CreatedAt.Unix(),         // 预期创建时间转换为 Unix 时间戳
		TOTPRegistered: user.TOTPRegistered,           // 预期 TOTP 状态保持不变
		RecoveryCode:   user.RecoveryCode,             // 预期恢复码保持不变
	}

	var result UserJSON // 用于存储 JSON 解码后的结果

	// 调用被测试对象的 EncodeToJSON 方法，获取 JSON 字符串
	jsonString := user.EncodeToJSON()
	// 将 JSON 字符串解码到 result 结构体中
	err := json.Unmarshal([]byte(jsonString), &result)
	assert.NoError(t, err) // 断言解码过程中没有错误

	// 断言解码后的结果 (result) 与预期的结果 (expected) 完全一致
	assert.Equal(t, expected, result)
}

// TestEncodeRecoveryCodeToJSON 测试 encodeRecoveryCodeToJSON 函数的功能。
// 这个函数 (推测定义在 user.go 或类似文件中) 专门用于将恢复码编码成一个简单的 JSON 对象。
//
// 测试步骤：
// 1. 定义一个恢复码字符串。
// 2. 定义预期的 JSON 输出结构 (RecoveryCodeJSON)，只包含 recovery_code 字段。
// 3. 调用 encodeRecoveryCodeToJSON() 获取 JSON 字符串。
// 4. 将 JSON 字符串解码回 RecoveryCodeJSON 结构体。
// 5. 断言解码后的结构体与预期结构体相等。
func TestEncodeRecoveryCodeToJSON(t *testing.T) {
	t.Parallel() // 允许与其他 Parallel 测试并行运行

	recoveryCode := "12345678" // 测试用的恢复码

	// 预期得到的 JSON 结构
	expected := RecoveryCodeJSON{
		RecoveryCode: recoveryCode,
	}

	var result RecoveryCodeJSON // 用于存储 JSON 解码后的结果

	// 调用被测试函数获取 JSON 字符串
	jsonString := encodeRecoveryCodeToJSON(recoveryCode)
	// 将 JSON 字符串解码到 result 结构体中
	err := json.Unmarshal([]byte(jsonString), &result)
	assert.NoError(t, err) // 断言解码过程中没有错误

	// 断言解码后的结果 (result) 与预期的结果 (expected) 完全一致
	assert.Equal(t, expected, result)
}

// UserJSON 是用于测试 User.EncodeToJSON() 方法的辅助结构体。
// 它定义了 User 对象在编码为 JSON 时应包含的公共字段及其格式。
// - Id: 用户唯一标识符。
// - CreatedAtUnix: 用户创建时间的 Unix 时间戳 (int64)。
// - RecoveryCode: 用户的恢复码，可能在某些流程中需要返回给用户。
// - TOTPRegistered: 标记用户是否已注册 TOTP (布尔值)。
// 注意：此结构不包含敏感信息，如 PasswordHash。
type UserJSON struct {
	Id             string `json:"id"`             // 用户 ID，对应 JSON 中的 "id" 键
	CreatedAtUnix  int64  `json:"created_at"`     // 创建时间的 Unix 时间戳，对应 JSON 中的 "created_at" 键
	RecoveryCode   string `json:"recovery_code"`  // 恢复码，对应 JSON 中的 "recovery_code" 键
	TOTPRegistered bool   `json:"totp_registered"`// TOTP 注册状态，对应 JSON 中的 "totp_registered" 键
}

// RecoveryCodeJSON 是用于测试 encodeRecoveryCodeToJSON() 函数的辅助结构体。
// 它定义了一个非常简单的 JSON 结构，仅包含用户的恢复码。
// 这可能用于只需要返回恢复码的特定 API 端点。
type RecoveryCodeJSON struct {
	RecoveryCode string `json:"recovery_code"` // 恢复码，对应 JSON 中的 "recovery_code" 键
}
