package main

import (
	"database/sql" // 导入数据库 SQL 包，用于数据库操作
	"faroe/ratelimit" // 导入项目内部的 ratelimit 包，用于配置速率限制器
	"testing"      // 导入 Go 的测试包
	"time"         // 导入时间包，用于设置时间间隔
)

// initializeTestDB 函数用于初始化一个用于测试的内存 SQLite 数据库。
// 它创建一个内存数据库实例，并在其上执行 schema.sql 中定义的数据库结构。
// 这确保了每个测试都在一个干净、隔离的环境中运行，不会相互干扰，也不会影响生产数据库。
//
// 参数:
//   t (*testing.T): 测试框架提供的测试上下文对象，用于报告错误。
//
// 返回值:
//   *sql.DB: 初始化成功并应用了 schema 的内存数据库连接。
//            如果初始化或执行 schema 失败，则会调用 t.Fatal() 中止测试。
func initializeTestDB(t *testing.T) *sql.DB {
	// 使用 "sqlite" 驱动和 ":memory:" 数据源名称来创建内存数据库
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		// 如果打开数据库失败，记录致命错误并终止测试
		t.Fatal(err)
	}
	// 执行全局变量 schema 中定义的 SQL 语句 (通常是 CREATE TABLE 等)
	_, err = db.Exec(schema)
	if err != nil {
		// 如果执行 schema 失败，先关闭数据库连接，然后记录致命错误并终止测试
		db.Close()
		t.Fatal(err)
	}
	// 返回成功初始化的数据库连接
	return db
}

// createEnvironment 函数创建一个用于测试的 *Environment 实例。
// 它将测试数据库、一个测试用的密钥 (secret) 以及一系列配置好的速率限制器
// 注入到 Environment 结构体中。
// 这使得测试可以直接调用需要 Environment 依赖的函数，并控制这些依赖项的行为。
// 速率限制器的配置可能与生产环境不同，通常在测试中会设置得更宽松或具有更短的重置周期，
// 以便更容易地触发和测试限流逻辑，而无需等待很长时间。
//
// 参数:
//   db (*sql.DB):  已经初始化好的测试数据库连接 (通常来自 initializeTestDB)。
//   secret ([]byte): 用于测试的共享密钥，例如用于 JWT 或其他加密操作。
//
// 返回值:
//   *Environment: 配置了测试依赖项的 Environment 实例。
func createEnvironment(db *sql.DB, secret []byte) *Environment {
	// 初始化 Environment 结构体
	env := &Environment{
		db:                              db,      // 注入测试数据库
		secret:                          secret,  // 注入测试密钥
		// 初始化各种速率限制器，使用 ratelimit 包中的构造函数。
		// 注意：这里的参数 (如 max=5, interval=10*time.Second) 是为测试设置的，
		// 可能与生产环境配置不同，以便于测试。
		passwordHashingIPRateLimit:      ratelimit.NewTokenBucketRateLimit(5, 10*time.Second),       // 密码哈希 IP 速率限制 (补充型令牌桶)
		loginIPRateLimit:                ratelimit.NewExpiringTokenBucketRateLimit(5, 15*time.Minute), // 登录 IP 速率限制 (过期型令牌桶)
		createEmailRequestUserRateLimit: ratelimit.NewTokenBucketRateLimit(3, 5*time.Minute),        // 创建邮件请求用户速率限制 (补充型令牌桶)
		verifyUserEmailRateLimit:        ratelimit.NewExpiringTokenBucketRateLimit(5, 15*time.Minute), // 验证用户邮箱速率限制 (过期型令牌桶)
		verifyEmailUpdateVerificationCodeLimitCounter: ratelimit.NewLimitCounter(5),                   // 验证邮箱更新验证码次数限制 (计数器)
		createPasswordResetIPRateLimit:                ratelimit.NewTokenBucketRateLimit(3, 5*time.Minute),        // 创建密码重置 IP 速率限制 (补充型令牌桶)
		verifyPasswordResetCodeLimitCounter:           ratelimit.NewLimitCounter(5),                   // 验证密码重置码次数限制 (计数器)
		totpUserRateLimit:                             ratelimit.NewExpiringTokenBucketRateLimit(5, 15*time.Minute), // TOTP 用户速率限制 (过期型令牌桶)
		recoveryCodeUserRateLimit:                     ratelimit.NewExpiringTokenBucketRateLimit(5, 15*time.Minute), // 恢复码用户速率限制 (过期型令牌桶)
	}
	// 返回配置好的测试环境实例
	return env
}

// ErrorJSON 结构体用于在集成测试中解析 API 返回的 JSON 格式错误响应。
// 当测试需要验证 API 是否按预期返回了特定的错误信息时，可以将响应体 unmarshal 到这个结构体中，
// 然后检查 Error 字段的值。
type ErrorJSON struct {
	Error string `json:"error"` // 对应 JSON 中的 "error" 字段
}
