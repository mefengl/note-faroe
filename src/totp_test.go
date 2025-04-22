package main

import (
	"database/sql"    // 导入数据库 SQL 包
	"encoding/base64" // 导入 Base64 编码包，用于处理二进制密钥
	"encoding/json"   // 导入 JSON 编码/解码包
	"testing"         // 导入 Go 的测试包
	"time"            // 导入时间包

	"github.com/stretchr/testify/assert" // 导入 testify 断言库
)

// insertUserTOTPCredential 是一个测试辅助函数，用于向数据库中插入一条用户 TOTP (基于时间的一次性密码) 凭证记录。
// 这通常在需要预设 TOTP 数据进行其他测试时使用。
// 参数：
//   db (*sql.DB): 数据库连接对象。
//   credential (*UserTOTPCredential): 要插入的 TOTP 凭证数据。
// 返回值：
//   error: 如果数据库操作出错，则返回错误信息，否则返回 nil。
func insertUserTOTPCredential(db *sql.DB, credential *UserTOTPCredential) error {
	// 执行 SQL INSERT 语句，将用户 ID、创建时间 (Unix 时间戳) 和 TOTP 密钥插入到 user_totp_credential 表中。
	// Key 是 []byte 类型，直接存储在数据库中（具体存储方式取决于数据库和驱动）。
	_, err := db.Exec("INSERT INTO user_totp_credential (user_id, created_at, key) VALUES (?, ?, ?)", credential.UserId, credential.CreatedAt.Unix(), credential.Key)
	return err // 返回执行结果的错误信息 (如果存在)
}

// TestUserTOTPCredentialEncodeToJSON 测试 UserTOTPCredential 结构体的 EncodeToJSON 方法。
// 这个测试主要验证当 UserTOTPCredential 对象被编码为 JSON 时：
// 1. UserId 和 CreatedAt (转换为 Unix 时间戳) 字段被正确包含。
// 2. Key ([]byte 类型) 字段被正确地进行 Base64 编码后包含在 JSON 中 (通常是为了方便传输和显示)。
//
// 测试步骤：
// 1. 创建一个 UserTOTPCredential 实例，包含用户 ID、创建时间和二进制密钥。
// 2. 定义预期的 JSON 输出结构 (UserTOTPCredentialJSON)，其中密钥字段 (EncodedKey) 应为原始密钥的 Base64 编码字符串。
// 3. 调用 credential.EncodeToJSON() 获取 JSON 字符串。
// 4. 将返回的 JSON 字符串解码回 UserTOTPCredentialJSON 结构体。
// 5. 使用 assert.Equal 断言解码后的结构体与预期的结构体完全相等。
func TestUserTOTPCredentialEncodeToJSON(t *testing.T) {
	t.Parallel() // 允许此测试与其他 Parallel 测试并行运行

	// 获取当前时间并截断纳秒，用于创建时间戳
	now := time.Unix(time.Now().Unix(), 0)

	// 创建一个测试用的 UserTOTPCredential 实例
	credential := UserTOTPCredential{
		UserId:    "1",                           // 用户 ID
		CreatedAt: now,                           // 创建时间
		Key:       []byte{0x01, 0x02, 0x03},      // 一个简单的二进制密钥 (byte 切片)
	}

	// 预期得到的 JSON 结构。注意 Key 字段被 Base64 编码为字符串。
	expected := UserTOTPCredentialJSON{
		UserId:        credential.UserId,                 // 预期用户 ID 保持不变
		CreatedAtUnix: credential.CreatedAt.Unix(),       // 预期创建时间转换为 Unix 时间戳
		EncodedKey:    base64.StdEncoding.EncodeToString(credential.Key), // 预期密钥被 Base64 编码
	}

	var result UserTOTPCredentialJSON // 用于存储 JSON 解码后的结果

	// 调用被测试对象的 EncodeToJSON 方法，获取 JSON 字符串
	jsonString := credential.EncodeToJSON()
	// 将 JSON 字符串解码到 result 结构体中
	err := json.Unmarshal([]byte(jsonString), &result)
	assert.NoError(t, err) // 断言解码过程中没有错误发生

	// 断言解码后的结果 (result) 与预期的结果 (expected) 完全一致
	assert.Equal(t, expected, result)
}

// UserTOTPCredentialJSON 是用于在测试中表示 UserTOTPCredential 编码为 JSON 后的预期结构。
// 它定义了 JSON 输出应包含的字段及其类型。
// 特别注意，原始的 []byte 类型的 Key 在这里表示为 Base64 编码的字符串 EncodedKey。
type UserTOTPCredentialJSON struct {
	UserId        string `json:"user_id"`    // 用户 ID，对应 JSON 中的 "user_id" 键
	CreatedAtUnix int64  `json:"created_at"` // 创建时间的 Unix 时间戳，对应 JSON 中的 "created_at" 键
	EncodedKey    string `json:"key"`        // Base64 编码后的密钥字符串，对应 JSON 中的 "key" 键
}
