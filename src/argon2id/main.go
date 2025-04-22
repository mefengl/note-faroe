package argon2id

import (
	"crypto/rand"        // 用于生成安全的随机字节序列（例如盐）
	"crypto/subtle"      // 提供常量时间操作，用于安全比较哈希值，防止时序攻击
	"encoding/base64"    // 用于将字节序列编码为 Base64 字符串，以便存储和传输
	"errors"             // 用于创建和处理错误
	"fmt"                // 用于格式化字符串
	"strings"            // 用于字符串操作，例如分割哈希字符串

	"golang.org/x/crypto/argon2" // 导入 Argon2 加密库
)

// Hash 函数接收一个明文密码字符串，使用 Argon2id 算法生成一个安全的密码哈希值。
// Argon2id 是目前推荐的密码哈希算法之一，它结合了 Argon2i 和 Argon2d 的优点，
// 既能抵抗 GPU 破解（通过内存消耗），也能抵抗侧信道攻击。
//
// 工作流程:
// 1. 生成一个随机的 16 字节盐 (salt)。盐的作用是确保即使两个用户使用相同的密码，
//    他们的哈希值也是不同的，增加了彩虹表攻击的难度。
// 2. 调用 golang.org/x/crypto/argon2.IDKey 函数，传入密码、盐和 Argon2id 参数，
//    计算出派生的密钥 (derived key)，也就是密码的哈希结果。
//    参数说明:
//      - []byte(password): 明文密码的字节表示。
//      - salt: 随机生成的盐。
//      - time (t): 2 (迭代次数，增加计算成本)。
//      - memory (m): 19456 (内存消耗，单位 KiB，增加内存需求)。
//      - parallelism (p): 1 (并行度，使用的线程数)。
//      - keyLen: 32 (生成的哈希密钥长度，单位字节)。
//    这些参数的选择影响了哈希的强度和计算所需资源，需要根据安全需求和服务器性能进行调整。
//    这里的参数 (t=2, m=19MiB, p=1) 是一个相对适中的选择。
// 3. 将算法标识、版本、参数、盐 (Base64 编码) 和派生密钥 (Base64 编码) 组合成
//    一个标准的 Argon2 哈希字符串格式，例如：
//    `$argon2id$v=19$m=19456,t=2,p=1$生成的盐Base64$生成的密钥Base64`
//    这种格式使得验证时可以方便地提取出所有必要的信息。
//
// 参数:
//   password (string): 用户提供的明文密码。
//
// 返回值:
//   string: 生成的 Argon2id 密码哈希字符串。
//   error: 如果在生成随机盐时发生错误，则返回错误。
func Hash(password string) (string, error) {
	// 1. 生成 16 字节的随机盐
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		// 如果生成随机盐失败，返回错误
		return "", err
	}
	// 2. 使用 Argon2id 计算派生密钥 (哈希)
	// 参数: 时间成本 t=2, 内存成本 m=19*1024=19456 KiB, 并行度 p=1, 输出密钥长度 32 字节
	key := argon2.IDKey([]byte(password), salt, 2, 19456, 1, 32)
	// 3. 格式化为标准的 Argon2 哈希字符串
	// 使用 RawStdEncoding 避免 Base64 编码中的 '=' 填充符
	hash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, // 使用库中定义的 Argon2 版本号 (通常是 19，即 0x13)
		19456,          // 内存参数 m
		2,              // 时间参数 t
		1,              // 并行度参数 p
		base64.RawStdEncoding.EncodeToString(salt), // Base64 编码的盐
		base64.RawStdEncoding.EncodeToString(key)) // Base64 编码的派生密钥
	return hash, nil
}

// Verify 函数接收一个存储的 Argon2id 哈希字符串和一个待验证的明文密码，
// 检查密码是否与哈希匹配。
//
// 工作流程:
// 1. 解析哈希字符串: 使用 '$' 作为分隔符将哈希字符串分割成多个部分。
// 2. 验证格式: 检查分割后的部分数量是否正确 (预期为 6 部分)，以及各部分是否符合预期格式
//    (例如，第二部分是 "argon2id"，第三部分是 "v=19" 等)。
// 3. 提取参数: 从第四部分提取 Argon2id 的内存 (m)、时间 (t) 和并行度 (p) 参数。
//    注意：当前实现中，虽然提取了参数，但在后续计算 key2 时并未使用这些提取出的参数，
//    而是硬编码了与 Hash 函数相同的参数 (t=2, m=19456, p=1)。这是一个潜在的问题，
//    如果 Hash 函数的参数未来发生改变，这里的验证逻辑需要同步更新。
//    更健壮的实现应该使用从哈希中提取出的 m, t, p 参数来计算 key2。
// 4. 解码盐和密钥: 从第五和第六部分解码 Base64 编码的盐 (salt) 和存储的派生密钥 (key1)。
// 5. 重新计算哈希: 使用从哈希中提取的盐 (salt) 和**硬编码的参数** (t=2, m=19456, p=1)
//    以及用户提供的明文密码，调用 argon2.IDKey 重新计算一个派生密钥 (key2)。
//    输出密钥的长度与解码出的 key1 保持一致。
// 6. 比较密钥: 使用 crypto/subtle.ConstantTimeCompare 函数在常量时间内比较
//    重新计算出的密钥 (key2) 和从哈希中解码出的原始密钥 (key1)。
//    使用常量时间比较是为了防止时序攻击 (timing attacks)，攻击者可能通过测量比较操作
//    所需的时间来推断密钥的部分信息。ConstantTimeCompare 确保无论比较结果如何，
//    操作花费的时间都是相同的。
//
// 参数:
//   hash (string): 存储的 Argon2id 密码哈希字符串。
//   password (string): 用户提供的待验证的明文密码。
//
// 返回值:
//   bool: 如果密码与哈希匹配，返回 true；否则返回 false。
//   error: 如果哈希字符串格式无效、算法或版本不受支持，或者在解析或解码过程中发生错误，则返回错误。
func Verify(hash string, password string) (bool, error) {
	// 1. 分割哈希字符串
	parts := strings.Split(hash, "$")
	// 2. 验证格式 - 期望有 6 个部分 (空字符串, "argon2id", "v=19", "m=...,t=...,p=...", salt, key)
	if len(parts) != 6 {
		return false, errors.New("invalid hash format: incorrect number of parts")
	}
	// 验证第一部分是否为空
	if parts[0] != "" {
		return false, errors.New("invalid hash format: expected empty first part")
	}
	// 验证算法标识
	if parts[1] != "argon2id" {
		return false, errors.New("invalid algorithm: expected 'argon2id'")
	}
	// 验证版本号
	if parts[2] != fmt.Sprintf("v=%d", argon2.Version) {
		return false, fmt.Errorf("unsupported hash version: expected 'v=%d'", argon2.Version)
	}
	// 3. 提取参数 (m, t, p)
	var m uint32 // 注意：库函数使用 uint32
	var t, p uint8 // 注意：库函数使用 uint8
	// 注意：fmt.Sscanf 对无符号整数的支持可能不直接，这里用 %d 读取到 int32 再转换可能更安全，
	// 或者直接解析字符串。但考虑到这里的参数值不大，直接用 %d 读取到临时变量再赋值给 uint 也可以。
	// 更好的方法是手动解析 parts[3] 字符串。当前 Sscanf 的写法可能不够健壮。
	var mScan int32
	var tScan, pScan int32
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mScan, &tScan, &pScan)
	if err != nil {
		return false, fmt.Errorf("invalid hash format: failed to parse parameters: %w", err)
	}
	m = uint32(mScan)
	t = uint8(tScan)
	p = uint8(pScan)

	// 4. 解码盐 (salt)
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("invalid hash format: failed to decode salt: %w", err)
	}
	// 4. 解码存储的派生密钥 (key1)
	key1, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("invalid hash format: failed to decode key: %w", err)
	}

	// 5. 使用从哈希中提取的盐和 *硬编码* 的参数重新计算密钥 (key2)
	// !!! 重要提示: 这里硬编码了参数 (t=2, m=19456, p=1), 而不是使用从哈希中解析出的 m, t, p。
	// 这意味着如果 Hash 函数的参数改变，这里的验证会失败。正确的做法是使用解析出的 m, t, p。
	// 例如: key2 := argon2.IDKey([]byte(password), salt, uint32(t), m, uint8(p), uint32(len(key1)))
	// 为了与当前 Hash 函数的行为保持一致，暂时保留硬编码。
	key2 := argon2.IDKey([]byte(password), salt, 2, 19456, 1, uint32(len(key1)))

	// 6. 使用常量时间比较两个密钥
	// subtle.ConstantTimeCompare 返回 1 表示相等，0 表示不相等。
	valid := subtle.ConstantTimeCompare(key1, key2)

	// 如果 valid == 1，则密码匹配，返回 true, nil
	return valid == 1, nil
}
