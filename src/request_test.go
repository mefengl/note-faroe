package main

import (
	"net/http/httptest" // 导入 httptest 包，用于创建模拟的 HTTP 请求对象
	"testing"          // 导入 Go 的测试包

	"github.com/stretchr/testify/assert" // 导入 testify 断言库，用于进行测试断言
)

// TestVerifyRequestSecret 测试 verifyRequestSecret 函数的功能。
// 这个测试的目的是验证 verifyRequestSecret 函数是否能够正确地根据服务器配置的密钥 (secret)
// 来检查传入 HTTP 请求的 "Authorization" 头部信息。
//
// 测试场景包括:
// 1. 服务器未配置密钥 (secret 为空字节切片):
//    - 请求包含 "Authorization" 头: 应该验证通过 (返回 true)。
//    - 请求不包含 "Authorization" 头或头为空: 应该验证通过 (返回 true)。
// 2. 服务器配置了密钥 (secret 不为空):
//    - 请求包含与服务器密钥完全匹配的 "Authorization" 头: 应该验证通过 (返回 true)。
//    - 请求不包含 "Authorization" 头或头为空: 应该验证失败 (返回 false)。
//    - 请求包含 "Authorization" 头，但与服务器密钥不匹配: (此场景未显式测试，但隐含在逻辑中，也会失败)
//    - 请求对象本身没有设置 Header (例如 Header 为 nil): 应该验证失败 (返回 false)。
func TestVerifyRequestSecret(t *testing.T) {
	// 场景 1.1: 服务器 secret 为空，请求头有 Authorization
	r := httptest.NewRequest("GET", "/", nil) // 创建一个模拟 GET 请求
	r.Header.Set("Authorization", "abc")      // 设置 Authorization 头
	// 断言：当服务器 secret 为空时，无论请求 Authorization 是什么，都应返回 true
	assert.Equal(t, true, verifyRequestSecret([]byte{}, r))

	// 场景 1.2: 服务器 secret 为空，请求头 Authorization 为空
	r = httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "") // 设置空的 Authorization 头
	// 断言：当服务器 secret 为空时，即使请求 Authorization 为空，也应返回 true
	assert.Equal(t, true, verifyRequestSecret([]byte{}, r))

	// 场景 2.1: 服务器 secret 非空，请求头 Authorization 匹配
	r = httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "abc") // 设置与服务器 secret 匹配的 Authorization 头
	// 断言：当服务器 secret 非空且请求 Authorization 匹配时，应返回 true
	assert.Equal(t, true, verifyRequestSecret([]byte("abc"), r))

	// 场景 2.2: 服务器 secret 非空，请求头 Authorization 为空
	r = httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "") // 设置空的 Authorization 头
	// 断言：当服务器 secret 非空但请求 Authorization 为空时，应返回 false
	assert.Equal(t, false, verifyRequestSecret([]byte("abc"), r))

	// 场景 2.3: 服务器 secret 非空，请求没有 Authorization 头 (Header 存在但 Key 不存在)
	r = httptest.NewRequest("GET", "/", nil) // 创建请求，不设置 Authorization 头
	// 断言：当服务器 secret 非空但请求缺少 Authorization 头时，应返回 false
	assert.Equal(t, false, verifyRequestSecret([]byte("abc"), r))

	// 注意：verifyRequestSecret 函数内部可能还会处理 r.Header 为 nil 的情况，
	// 但此测试用例没有显式覆盖 r.Header 本身就是 nil 的场景。
	// httptest.NewRequest 总是会初始化 Header。
}
