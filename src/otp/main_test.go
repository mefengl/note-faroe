package otp

import (
	"fmt"
	"testing" // 导入 Go 的测试包
)

// TestGenerateHOTP 测试 GenerateHOTP 函数的正确性。
// GenerateHOTP 用于基于 HMAC (Hash-based Message Authentication Code) 生成一次性密码。
// 它需要一个密钥 (key)，一个计数器 (counter)，以及期望的密码长度 (digits)。
//
// 测试步骤：
// 1. 定义一个固定的测试密钥 (key)，这里使用全 0xff 的字节数组。
// 2. 定义一组测试用例 (tests)，每个用例包含一个计数器值 (counter) 和对应的预期 HOTP 值 (expected)。
//    这些预期值通常来自 RFC 4226 的附录或其他标准参考实现。
// 3. 遍历测试用例，为每个用例创建一个子测试 (t.Run)。
// 4. 在子测试中，调用 GenerateHOTP 函数，传入密钥、当前测试用例的计数器和固定的密码长度 (6)。
// 5. 将生成的 HOTP 结果 (result) 与当前测试用例的预期值 (test.expected) 进行比较。
// 6. 如果结果与预期不符，则通过 t.Errorf 报告错误。
func TestGenerateHOTP(t *testing.T) {
	// 创建一个 20 字节的密钥，并用 0xff 填充
	key := make([]byte, 20)
	for i := 0; i < len(key); i++ {
		key[i] = 0xff
	}

	// 定义一系列测试用例，包含不同的计数器值及其预期的 6 位 HOTP 结果
	tests := []struct {
		counter  uint64 // 计数器
		expected string // 预期的 HOTP 字符串
	}{
		{0, "103905"},
		{1, "463444"},
		{10, "413510"},
		{100, "632126"},
		{10000, "529078"},
		{100000000, "818472"},
	}

	// 遍历所有测试用例
	for _, test := range tests {
		// 为每个计数器创建一个子测试，方便定位问题
		t.Run(fmt.Sprintf("Counter: %d", test.counter), func(t *testing.T) {
			// 调用 GenerateHOTP 函数生成实际的 HOTP
			result := GenerateHOTP(key, test.counter, 6) // 生成 6 位密码
			// 比较实际结果与预期结果
			if result != test.expected {
				// 如果不匹配，报告错误
				t.Errorf("got %s, expected %s", result, test.expected)
			}
		})
	}
}

// TestVerifyHOTP 测试 VerifyHOTP 函数的正确性。
// VerifyHOTP 用于验证用户提供的一次性密码 (otp) 是否与基于密钥和计数器计算出的密码匹配。
//
// 测试步骤：
// 1. 定义与 TestGenerateHOTP 中相同的测试密钥 (key)。
// 2. 定义一组有效的测试用例 (validTests)，包含计数器和对应的正确 HOTP 值。
// 3. 定义一组无效的测试用例 (invalidTests)，包含计数器和错误的 HOTP 值。
// 4. 遍历有效的测试用例：
//    a. 为每个用例创建子测试。
//    b. 调用 VerifyHOTP 函数，传入密钥、计数器、密码长度和正确的 OTP。
//    c. 断言 VerifyHOTP 应返回 true (验证通过)。如果返回 false，则报告错误。
// 5. 遍历无效的测试用例：
//    a. 为每个用例创建子测试。
//    b. 调用 VerifyHOTP 函数，传入密钥、计数器、密码长度和错误的 OTP。
//    c. 断言 VerifyHOTP 应返回 false (验证失败)。如果返回 true，则报告错误。
func TestVerifyHOTP(t *testing.T) {
	// 创建与生成测试中相同的密钥
	key := make([]byte, 20)
	for i := 0; i < len(key); i++ {
		key[i] = 0xff
	}

	// 定义有效的测试用例（计数器和对应的正确 OTP）
	validTests := []struct {
		counter uint64 // 计数器
		otp     string // 正确的 OTP
	}{
		{0, "103905"},
		{1, "463444"},
		{10, "413510"},
		{100, "632126"},
		{10000, "529078"},
		{100000000, "818472"},
	}

	// 定义无效的测试用例（例如，OTP 最后一位错误）
	invalidTests := []struct {
		counter uint64 // 计数器
		otp     string // 错误的 OTP
	}{
		{0, "103906"}, // OTP 与 counter 0 的预期值 "103905" 不符
	}

	// 遍历并测试所有有效的 OTP
	for _, test := range validTests {
		t.Run(fmt.Sprintf("Valid Counter: %d", test.counter), func(t *testing.T) {
			// 使用正确的 OTP 调用 VerifyHOTP
			result := VerifyHOTP(key, test.counter, 6, test.otp) // 验证 6 位密码
			// 预期结果应为 true (验证成功)
			if !result {
				t.Error("got false, expected true") // 如果失败，报告错误
			}
		})
	}

	// 遍历并测试所有无效的 OTP
	for _, test := range invalidTests {
		t.Run(fmt.Sprintf("Invalid Counter: %d", test.counter), func(t *testing.T) {
			// 使用错误的 OTP 调用 VerifyHOTP
			result := VerifyHOTP(key, test.counter, 6, test.otp) // 验证 6 位密码
			// 预期结果应为 false (验证失败)
			if result {
				t.Error("got true, expected false") // 如果成功，报告错误
			}
		})
	}
}
