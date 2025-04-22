package main

import (
	"encoding/json" // 导入 JSON 编码/解码包
	"testing"         // 导入 Go 的测试包
	"time"            // 导入时间包

	"github.com/stretchr/testify/assert" // 导入 testify 断言库
)

// TestPasswordResetRequestEncodeToJSON 测试 PasswordResetRequest 结构体的 EncodeToJSON 方法。
// 这个测试验证当调用 EncodeToJSON 时，是否只序列化了 Id, UserId, CreatedAt (转为 Unix 时间戳),
// 和 ExpiresAt (转为 Unix 时间戳) 这几个字段，而忽略了 CodeHash。
//
// 测试步骤：
// 1. 创建一个 PasswordResetRequest 实例。
// 2. 定义预期的 JSON 输出结构 (PasswordResetRequestJSON)，只包含上述四个字段。
// 3. 调用 request.EncodeToJSON() 获取 JSON 字符串。
// 4. 将 JSON 字符串解码回 PasswordResetRequestJSON 结构体。
// 5. 断言解码后的结构体与预期结构体相等。
func TestPasswordResetRequestEncodeToJSON(t *testing.T) {
	t.Parallel() // 允许与其他 Parallel 测试并行运行

	// 获取当前时间并截断纳秒，用于测试时间戳
	now := time.Unix(time.Now().Unix(), 0)

	// 创建一个测试用的 PasswordResetRequest 实例
	request := PasswordResetRequest{
		Id:        "1",
		UserId:    "1",
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute), // 过期时间设置为 10 分钟后
		CodeHash:  "HASH1",                   // 设置一个 CodeHash，验证它不会被序列化
	}

	// 预期得到的 JSON 结构，不包含 CodeHash，时间为 Unix 时间戳
	expected := PasswordResetRequestJSON{
		Id:            request.Id,
		UserId:        request.UserId,
		CreatedAtUnix: request.CreatedAt.Unix(),
		ExpiresAtUnix: request.ExpiresAt.Unix(),
	}

	var result PasswordResetRequestJSON // 用于存储解码后的结果

	// 调用 EncodeToJSON 方法，并将返回的 JSON 字符串解码到 result 中
	err := json.Unmarshal([]byte(request.EncodeToJSON()), &result)
	assert.NoError(t, err) // 断言解码过程没有错误

	// 断言解码后的结果与预期结果完全一致
	assert.Equal(t, expected, result)
}

// TestPasswordResetRequestEncodeToJSONWithCode 测试 PasswordResetRequest 结构体的 EncodeToJSONWithCode 方法。
// 这个测试验证当调用 EncodeToJSONWithCode 时，是否序列化了 Id, UserId, CreatedAt (Unix 时间戳),
// ExpiresAt (Unix 时间戳), 以及**传入的 code** 字段，同样忽略了结构体本身的 CodeHash。
//
// 测试步骤：
// 1. 创建一个 PasswordResetRequest 实例。
// 2. 定义一个临时的 code 字符串。
// 3. 定义预期的 JSON 输出结构 (PasswordResetRequestWithCodeJSON)，包含基本字段和传入的 code。
// 4. 调用 request.EncodeToJSONWithCode(code) 获取 JSON 字符串。
// 5. 将 JSON 字符串解码回 PasswordResetRequestWithCodeJSON 结构体。
// 6. 断言解码后的结构体与预期结构体相等。
func TestPasswordResetRequestEncodeToJSONWithCode(t *testing.T) {
	t.Parallel() // 允许与其他 Parallel 测试并行运行

	// 获取当前时间并截断纳秒
	now := time.Unix(time.Now().Unix(), 0)

	code := "12345678" // 定义一个要包含在 JSON 中的明文 code
	// 创建一个测试用的 PasswordResetRequest 实例
	request := PasswordResetRequest{
		Id:        "1",
		UserId:    "1",
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute), // 过期时间设置为 10 分钟后
		CodeHash:  "HASH1",                   // 设置 CodeHash，验证它仍然被忽略
	}

	// 预期得到的 JSON 结构，包含基本字段和传入的 code
	expected := PasswordResetRequestWithCodeJSON{
		Id:            request.Id,
		UserId:        request.UserId,
		CreatedAtUnix: request.CreatedAt.Unix(),
		ExpiresAtUnix: request.ExpiresAt.Unix(),
		Code:          code, // 预期 JSON 中包含传入的 code
	}

	var result PasswordResetRequestWithCodeJSON // 用于存储解码后的结果

	// 调用 EncodeToJSONWithCode 方法，传入 code，并将返回的 JSON 字符串解码到 result 中
	err := json.Unmarshal([]byte(request.EncodeToJSONWithCode(code)), &result)
	assert.NoError(t, err) // 断言解码过程没有错误

	// 断言解码后的结果与预期结果完全一致
	assert.Equal(t, expected, result)
}

// PasswordResetRequestJSON 是用于测试 PasswordResetRequest.EncodeToJSON() 方法的辅助结构体。
// 它定义了预期的 JSON 输出格式，只包含基本的请求信息，不含敏感的哈希值或明文代码。
// 时间字段使用 Unix 时间戳表示。
type PasswordResetRequestJSON struct {
	Id            string `json:"id"`         // 请求 ID，对应 JSON 中的 "id" 键
	UserId        string `json:"user_id"`    // 用户 ID，对应 JSON 中的 "user_id" 键
	CreatedAtUnix int64  `json:"created_at"` // 创建时间的 Unix 时间戳，对应 JSON 中的 "created_at" 键
	ExpiresAtUnix int64  `json:"expires_at"` // 过期时间的 Unix 时间戳，对应 JSON 中的 "expires_at" 键
}

// PasswordResetRequestWithCodeJSON 是用于测试 PasswordResetRequest.EncodeToJSONWithCode() 方法的辅助结构体。
// 它定义了包含明文重置代码 (Code) 的 JSON 输出格式。这通常用于向用户显示一次性的重置代码。
// 时间字段同样使用 Unix 时间戳表示。
type PasswordResetRequestWithCodeJSON struct {
	Id            string `json:"id"`         // 请求 ID
	UserId        string `json:"user_id"`    // 用户 ID
	CreatedAtUnix int64  `json:"created_at"` // 创建时间的 Unix 时间戳
	ExpiresAtUnix int64  `json:"expires_at"` // 过期时间的 Unix 时间戳
	Code          string `json:"code"`       // 明文重置代码，对应 JSON 中的 "code" 键
}
