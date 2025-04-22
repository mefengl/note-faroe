package otp

import (
	"crypto/hmac"      // 用于计算 HMAC (Hash-based Message Authentication Code)
	"crypto/sha1"      // 使用 SHA1 作为 HMAC 的哈希函数 (注意: SHA1 已不推荐用于新应用，但 TOTP/HOTP 标准仍常用)
	"crypto/subtle"      // 提供常量时间比较函数，防止时序攻击
	"encoding/binary"  // 用于在字节序列和数值类型之间进行转换 (大端序)
	"math"             // 用于数学计算，例如计算 10 的幂次方
	"strconv"          // 用于字符串和基本数据类型之间的转换
	"time"             // 用于处理时间相关的操作
)

// GenerateTOTP 函数根据 RFC 6238 生成一个基于时间的一次性密码 (TOTP)。
// TOTP 是 HOTP 的一个变种，它使用当前时间除以时间间隔得到的整数作为计数器。
//
// 工作流程:
// 1. 计算当前时间戳 (Unix 秒数) 除以时间间隔 (秒数) 的整数部分，得到时间步长计数器。
// 2. 调用 GenerateHOTP 函数，传入共享密钥、计算出的计数器和指定的位数，生成最终的 OTP。
//
// 参数:
//   now (time.Time):       当前时间。
//   key ([]byte):          共享密钥 (通常是 Base32 解码后的字节)。
//   interval (time.Duration): 时间间隔，定义了 OTP 的有效期 (例如 30 秒)。
//   digits (int):          生成的 OTP 的位数 (通常是 6 或 8)。
//
// 返回值:
//   string: 生成的 TOTP 字符串 (例如 "123456")。
func GenerateTOTP(now time.Time, key []byte, interval time.Duration, digits int) string {
	// 计算时间步长计数器 (counter) = floor(当前 Unix 时间戳 / 时间间隔秒数)
	counter := uint64(now.Unix()) / uint64(interval.Seconds())
	// 调用 HOTP 生成函数，使用计算出的计数器
	return GenerateHOTP(key, counter, digits)
}

// VerifyTOTP 函数验证用户提供的 TOTP 是否在当前时间步长内有效。
//
// 工作流程:
// 1. 检查用户提供的 OTP 字符串长度是否与期望的位数匹配。
// 2. 调用 GenerateTOTP 函数生成当前时间步长的预期 OTP。
// 3. 使用 crypto/subtle.ConstantTimeCompare 函数在常量时间内比较生成的 OTP 和用户提供的 OTP。
//    这可以防止时序攻击，攻击者无法通过测量比较时间来猜测 OTP。
//
// 参数:
//   now (time.Time):       当前时间。
//   key ([]byte):          共享密钥。
//   interval (time.Duration): 时间间隔。
//   digits (int):          OTP 的位数。
//   otp (string):          用户提供的待验证的 OTP 字符串。
//
// 返回值:
//   bool: 如果 OTP 有效，返回 true；否则返回 false。
func VerifyTOTP(now time.Time, key []byte, interval time.Duration, digits int, otp string) bool {
	// 1. 检查 OTP 长度是否正确
	if len(otp) != digits {
		return false
	}
	// 2. 生成当前时间步长的预期 OTP
	generated := GenerateTOTP(now, key, interval, digits)
	// 3. 使用常量时间比较
	valid := subtle.ConstantTimeCompare([]byte(generated), []byte(otp)) == 1
	return valid
}

// VerifyTOTPWithGracePeriod 函数验证用户提供的 TOTP 是否在当前时间步长或其前后一个步长 (宽限期) 内有效。
// 这允许一定的时钟漂移或网络延迟。
//
// 工作流程:
// 1. 计算前一个时间步长 (now - gracePeriod) 的计数器，并生成对应的 OTP 进行比较。
// 2. 如果不匹配，计算当前时间步长 (now) 的计数器 (如果与前一个不同)，并生成对应的 OTP 进行比较。
// 3. 如果仍不匹配，计算后一个时间步长 (now + gracePeriod) 的计数器 (如果与前两个都不同)，并生成对应的 OTP 进行比较。
// 4. 只要在任何一个允许的时间步长内匹配成功，即返回 true。
// 5. 所有步长都验证失败，则返回 false。
// 注意: 这里 gracePeriod 通常应设置为等于 interval，以检查前一个、当前和下一个时间窗口。
//
// 参数:
//   now (time.Time):       当前时间。
//   key ([]byte):          共享密钥。
//   interval (time.Duration): 时间间隔。
//   digits (int):          OTP 的位数。
//   otp (string):          用户提供的待验证的 OTP 字符串。
//   gracePeriod (time.Duration): 允许的时间宽限期 (通常等于 interval)。
//
// 返回值:
//   bool: 如果 OTP 在宽限期内有效，返回 true；否则返回 false。
func VerifyTOTPWithGracePeriod(now time.Time, key []byte, interval time.Duration, digits int, otp string, gracePeriod time.Duration) bool {
	// 1. 检查前一个时间步长
	counter1 := uint64(now.Add(-1*gracePeriod).Unix()) / uint64(interval.Seconds())
	generated1 := GenerateHOTP(key, counter1, digits)
	valid1 := subtle.ConstantTimeCompare([]byte(generated1), []byte(otp)) == 1
	if valid1 {
		return true
	}

	// 2. 检查当前时间步长 (如果与前一个不同)
	counter2 := uint64(now.Unix()) / uint64(interval.Seconds())
	if counter2 != counter1 {
		generated2 := GenerateHOTP(key, counter2, digits)
		valid2 := subtle.ConstantTimeCompare([]byte(generated2), []byte(otp)) == 1
		if valid2 {
			return true
		}
	}

	// 3. 检查后一个时间步长 (如果与前两个都不同)
	counter3 := uint64(now.Add(gracePeriod).Unix()) / uint64(interval.Seconds())
	if counter3 != counter1 && counter3 != counter2 {
		generated3 := GenerateHOTP(key, counter3, digits)
		valid3 := subtle.ConstantTimeCompare([]byte(generated3), []byte(otp)) == 1
		if valid3 {
			return true
		}
	}

	// 4. 所有步长都验证失败
	return false
}

// GenerateHOTP 函数根据 RFC 4226 生成一个基于 HMAC 的一次性密码 (HOTP)。
// HOTP 是许多双因素认证系统的基础，包括 TOTP。
//
// 工作流程:
// 1. 验证位数是否在 6 到 8 之间 (标准要求)。
// 2. 将 64 位计数器 (counter) 转换为 8 字节的大端序字节序列。
// 3. 使用 HMAC-SHA1 算法计算计数器字节序列的 MAC (消息认证码)，密钥为共享密钥。
// 4. 对生成的 HMAC 结果 (hs，通常是 20 字节的 SHA1 哈希) 进行动态截断 (Dynamic Truncation):
//    a. 取 HMAC 结果的最后一个字节 (hs[19])。
//    b. 取该字节的低 4 位 (hs[19] & 0x0f)，这得到一个 0 到 15 之间的偏移量 (offset)。
//    c. 从 HMAC 结果中选取从 offset 开始的 4 个字节 (hs[offset : offset+4])。
// 5. 将这 4 个字节视为一个大端序的 32 位无符号整数 (snum)，但需要将最高位清零 (truncated[0] &= 0x7f)，
//    以确保结果是一个正整数，并避免符号位问题。
// 6. 计算该 32 位整数对 10^digits 取模的结果 (d = snum % 10^digits)。这会得到一个 0 到 10^digits - 1 之间的数。
// 7. 将结果 d 转换为字符串。
// 8. 如果字符串长度小于指定的位数 (digits)，在前面补零，直到达到指定长度。
// 9. 返回最终的 HOTP 字符串。
//
// 参数:
//   key ([]byte):     共享密钥 (通常是 Base32 解码后的字节)。
//   counter (uint64): 事件计数器或时间步长计数器。
//   digits (int):     生成的 OTP 的位数 (通常是 6 或 8)。
//
// 返回值:
//   string: 生成的 HOTP 字符串。
func GenerateHOTP(key []byte, counter uint64, digits int) string {
	// 1. 验证位数
	if digits < 6 || digits > 8 {
		// 根据 RFC 4226，位数通常是 6-8 位
		panic("invalid hotp digits: must be between 6 and 8")
	}
	// 2. 将计数器转为 8 字节大端序
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, counter)

	// 3. 计算 HMAC-SHA1
	mac := hmac.New(sha1.New, key) // 创建一个新的 HMAC 实例，使用 SHA1 和提供的密钥
	mac.Write(counterBytes)        // 写入要计算 HMAC 的数据 (计数器字节)
	hs := mac.Sum(nil)             // 计算并获取 HMAC 结果 (20 字节)

	// 4. 动态截断
	// a. 获取偏移量 (取哈希结果最后一个字节的低 4 位)
	offset := hs[len(hs)-1] & 0x0f
	// b. 提取 4 字节
	truncated := hs[offset : offset+4]

	// 5. 将 4 字节转为 32 位无符号整数，并清除最高位
	truncated[0] &= 0x7f // 清除最高位，确保结果为正数
	snum := binary.BigEndian.Uint32(truncated) // 按大端序解析为 uint32

	// 6. 计算模数
	// 计算 10 的 digits 次方 (10^digits)
	mod := uint32(math.Pow10(digits))
	// 取模得到最终数值
	d := snum % mod

	// 7. 格式化为字符串
	otp := strconv.Itoa(int(d)) // 将整数 d 转换为字符串

	// 8. 补零
	for len(otp) < digits {
		otp = "0" + otp
	}

	// 9. 返回 OTP
	return otp
}

// VerifyHOTP 函数验证用户提供的 HOTP 是否与给定计数器生成的 HOTP 匹配。
// 注意：HOTP 的验证通常需要同步计数器，这比 TOTP 更复杂。
// 这个函数本身只是简单地重新生成一次 HOTP 并进行比较。
//
// 参数:
//   key ([]byte):     共享密钥。
//   counter (uint64): 用于验证的计数器值。
//   digits (int):     OTP 的位数。
//   otp (string):     用户提供的待验证的 HOTP 字符串。
//
// 返回值:
//   bool: 如果 OTP 匹配，返回 true；否则返回 false。
func VerifyHOTP(key []byte, counter uint64, digits int, otp string) bool {
	// 生成预期的 HOTP 并直接与用户提供的 OTP 比较
	// 注意：这里没有使用常量时间比较，因为 HOTP 的验证场景通常不涉及对用户输入的直接反馈循环。
	// 但如果用于类似 TOTP 的场景，也应考虑使用常量时间比较。
	return GenerateHOTP(key, counter, digits) == otp
}
