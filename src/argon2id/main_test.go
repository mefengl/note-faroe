package argon2id

import "testing" // 导入 Go 的测试包

// Test 函数用于测试 argon2id 包中的 Hash 和 Verify 函数的功能。
// 它执行以下步骤：
// 1. 使用 Hash 函数为明文密码 "123456" 生成一个 Argon2id 哈希值。
// 2. 使用 Verify 函数验证生成的哈希值与原始密码 "123456" 是否匹配。
// 3. 使用 Verify 函数验证生成的哈希值与错误的密码 "12345" 是否匹配。
// 4. 检查在哈希和验证过程中是否发生错误。
//
// 测试预期：
// - Hash 函数应成功生成哈希值，不返回错误。
// - 第一次 Verify 调用（使用正确密码）应返回 true（有效）且无错误。
// - 第二次 Verify 调用（使用错误密码）应返回 false（无效）且无错误。
// - 如果任何步骤出现错误或验证结果不符合预期，测试将通过 t.Fatal 失败。
func Test(t *testing.T) {
	// 1. 对密码 "123456" 进行哈希处理
	hash, err := Hash("123456")
	// 检查哈希过程中是否发生错误
	if err != nil {
		// 如果出错，记录错误并立即终止测试
		t.Fatal(err)
	}

	// 2. 使用正确的密码 "123456" 验证哈希值
	valid, err := Verify(hash, "123456")
	// 检查验证过程中是否发生错误
	if err != nil {
		// 如果出错，记录错误并立即终止测试
		t.Fatal(err)
	}
	// 检查验证结果是否为 true (有效)
	if !valid {
		// 如果密码有效但验证失败，记录错误信息并终止测试
		t.Fatalf("Expected hash to match")
	}

	// 3. 使用错误的密码 "12345" 验证哈希值
	valid, err = Verify(hash, "12345")
	// 检查验证过程中是否发生错误
	if err != nil {
		// 如果出错，记录错误并立即终止测试
		t.Fatal(err)
	}
	// 检查验证结果是否为 false (无效)
	if valid {
		// 如果密码无效但验证成功，记录错误信息并终止测试
		t.Fatalf("Expected hash to not match")
	}
}
