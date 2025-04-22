package main

import (
	"crypto/rand"      // 导入用于生成加密安全的随机数的包
	"encoding/base32" // 导入用于 Base32 编码的包
)

// generateSecureCode 函数生成一个安全的、短小的、便于人类阅读和输入的验证码或令牌。
// 这种码通常用于邮箱验证、密码重置、两步验证确认等场景。
// 返回值:
//   string: 生成的 Base32 编码字符串 (例如 "A3K8P")。
//   error: 如果在生成随机字节时发生错误，则返回错误。
// 工作原理:
// 1. 创建一个 5 字节的切片 (bytes)。选择 5 字节是因为 Base32 编码会将 5 字节 (40 位) 转换为 8 个字符，
//    这是一个相对适中的长度，既足够安全 (理论上有 32^8 种可能性)，又不会太长导致用户输入困难。
// 2. 使用 crypto/rand.Read 填充这个字节切片。crypto/rand 使用操作系统提供的加密安全的随机数源，
//    这对于生成不可预测的验证码至关重要，可以防止攻击者猜测或暴力破解。
// 3. 如果 rand.Read 返回错误 (虽然很少见，但可能发生，例如系统随机数源出问题)，则函数返回空字符串和错误。
// 4. 定义一个自定义的 Base32 编码器。标准的 Base32 编码包含数字 0, 1 和字母 O, I。
//    这些字符在某些字体下容易混淆 (0 vs O, 1 vs I)，为了提高用户体验，这里创建了一个新的编码表，
//    移除了这些易混淆的字符。编码表为 "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"。
// 5. 使用这个自定义的编码器将随机生成的 5 个字节 (bytes) 编码成一个 Base32 字符串。
// 6. 返回生成的 Base32 字符串和 nil 错误。
func generateSecureCode() (string, error) {
	// 创建一个长度为 5 的字节切片，用于存储随机字节
	bytes := make([]byte, 5)
	// 使用加密安全的随机数生成器填充字节切片
	_, err := rand.Read(bytes)
	// 如果生成随机数时出错，返回错误
	if err != nil {
		return "", err
	}
	// 使用自定义的 Base32 编码将字节转换为字符串
	// 自定义编码表移除了易混淆的字符 '0', 'O', '1', 'I'
	code := base32.NewEncoding("ABCDEFGHJKLMNPQRSTUVWXYZ23456789").EncodeToString(bytes)
	// 返回生成的验证码和 nil 错误
	return code, nil
}
