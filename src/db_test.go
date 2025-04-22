package main

import (
	"context"      // 导入上下文包，虽然在此测试中未显式使用 context 的超时或取消，但数据库操作函数可能需要它
	"testing"      // 导入 Go 的测试包
	"time"         // 导入时间包，用于处理时间相关的操作，如设置过期时间

	"github.com/stretchr/testify/assert" // 导入 testify 断言库，提供更丰富的断言方法
)

// TestCleanUpDatabase 测试 cleanUpDatabase 函数的功能。
// 这个测试的目的是验证 cleanUpDatabase 函数是否能够正确地从数据库中删除过期的
// 密码重置请求 (password_reset_request) 和用户邮箱验证请求 (user_email_verification_request)。
//
// 测试步骤:
// 1. 初始化一个干净的内存数据库。
// 2. 创建几个测试用户。
// 3. 为用户创建多个密码重置请求记录，其中一些已过期，一些未过期。
// 4. 为用户创建多个邮箱验证请求记录，同样包含已过期和未过期的记录。
// 5. 调用被测试的 cleanUpDatabase 函数。
// 6. 查询数据库，检查剩余的密码重置请求和邮箱验证请求的数量。
// 7. 使用断言库 (assert) 验证剩余记录的数量是否符合预期（即只有未过期的记录被保留）。
func TestCleanUpDatabase(t *testing.T) {
	// 1. 初始化测试数据库
	db := initializeTestDB(t)
	// 使用 defer 确保测试结束后数据库连接被关闭
	defer db.Close()

	// 获取当前时间，并将纳秒部分截断，以确保时间戳的一致性
	now := time.Unix(time.Now().Unix(), 0)

	// --- 设置测试数据 ---

	// 创建用户 1
	user1 := User{
		Id:             "1",
		CreatedAt:      now,
		PasswordHash:   "HASH", // 使用简单的占位符哈希
		RecoveryCode:   "12345678",
		TOTPRegistered: false,
	}
	err := insertUser(db, context.Background(), &user1) // 插入用户 1 到数据库
	if err != nil {
		t.Fatal(err) // 如果插入失败，终止测试
	}

	// 创建密码重置请求 1 (已过期)
	resetRequest1 := PasswordResetRequest{
		Id:        "1",
		UserId:    user1.Id,
		CreatedAt: now,
		ExpiresAt: now.Add(-10 * time.Minute), // 过期时间设置为 10 分钟前
		CodeHash:  "HASH",
	}
	err = insertPasswordResetRequest(db, context.Background(), &resetRequest1)
	if err != nil {
		t.Fatal(err)
	}

	// 创建密码重置请求 2 (未过期)
	resetRequest2 := PasswordResetRequest{
		Id:        "2",
		UserId:    user1.Id,
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute), // 过期时间设置为 10 分钟后
		CodeHash:  "HASH",
	}
	err = insertPasswordResetRequest(db, context.Background(), &resetRequest2)
	if err != nil {
		t.Fatal(err)
	}

	// 创建密码重置请求 3 (未过期)
	resetRequest3 := PasswordResetRequest{
		Id:        "3",
		UserId:    user1.Id,
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute), // 过期时间设置为 10 分钟后
		CodeHash:  "HASH",
	}
	err = insertPasswordResetRequest(db, context.Background(), &resetRequest3)
	if err != nil {
		t.Fatal(err)
	}

	// 创建用户 2
	user2 := User{
		Id:             "2",
		CreatedAt:      now,
		PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ", // 示例 Argon2 哈希
		RecoveryCode:   "12345678",
		TOTPRegistered: false,
	}
	err = insertUser(db, context.Background(), &user2)
	if err != nil {
		t.Fatal(err)
	}

	// 创建用户 3
	user3 := User{
		Id:             "3",
		CreatedAt:      now,
		PasswordHash:   "$argon2id$v=19$m=19456,t=2,p=1$enc5MDZrSElTSVE0ODdTSw$CS/AV+PQs08MhdeIrHhfmQ", // 示例 Argon2 哈希
		RecoveryCode:   "12345678",
		TOTPRegistered: false,
	}
	err = insertUser(db, context.Background(), &user3)
	if err != nil {
		t.Fatal(err)
	}

	// 创建邮箱验证请求 1 (未过期)
	verificationRequest1 := UserEmailVerificationRequest{
		UserId:    user1.Id,
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: now.Add(10 * time.Minute), // 过期时间设置为 10 分钟后
	}
	err = insertUserEmailVerificationRequest(db, &verificationRequest1)
	if err != nil {
		t.Fatal(err)
	}

	// 创建邮箱验证请求 2 (已过期)
	verificationRequest2 := UserEmailVerificationRequest{
		UserId:    user2.Id,
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: now.Add(-10 * time.Minute), // 过期时间设置为 10 分钟前
	}
	err = insertUserEmailVerificationRequest(db, &verificationRequest2)
	if err != nil {
		t.Fatal(err)
	}

	// 创建邮箱验证请求 3 (已过期)
	verificationRequest3 := UserEmailVerificationRequest{
		UserId:    user3.Id,
		CreatedAt: now,
		Code:      "12345678",
		ExpiresAt: now.Add(-10 * time.Minute), // 过期时间设置为 10 分钟前
	}
	err = insertUserEmailVerificationRequest(db, &verificationRequest3)
	if err != nil {
		t.Fatal(err)
	}

	// --- 执行被测试的函数 ---
	err = cleanUpDatabase(db) // 调用数据库清理函数
	if err != nil {
		t.Fatal(err) // 如果清理函数出错，终止测试
	}

	// --- 验证结果 ---

	// 验证密码重置请求的数量
	var passwordResetRequestCount int
	// 查询清理后 password_reset_request 表中的记录总数
	err = db.QueryRow("SELECT count(*) FROM password_reset_request").Scan(&passwordResetRequestCount)
	if err != nil {
		t.Fatal(err) // 如果查询失败，终止测试
	}
	// 断言：预期应该只剩下 2 个未过期的密码重置请求 (resetRequest2, resetRequest3)
	assert.Equal(t, 2, passwordResetRequestCount)

	// 验证邮箱验证请求的数量
	var emailVerificationRequestCount int
	// 查询清理后 user_email_verification_request 表中的记录总数
	err = db.QueryRow("SELECT count(*) FROM user_email_verification_request").Scan(&emailVerificationRequestCount)
	if err != nil {
		t.Fatal(err) // 如果查询失败，终止测试
	}
	// 断言：预期应该只剩下 1 个未过期的邮箱验证请求 (verificationRequest1)
	assert.Equal(t, 1, emailVerificationRequestCount)
}
