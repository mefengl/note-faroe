package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"faroe/otp" // 导入自定义的 otp 包，用于 TOTP 生成和验证
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

// handleRegisterTOTPRequest 处理用户注册 TOTP 两因素认证的 API 请求。
// 用户在启用 2FA 时，通常会扫描一个二维码（包含了密钥 Key），然后输入应用生成的当前 TOTP 验证码 (Code)。
// 此函数接收用户 ID、密钥（Base64 编码）和用户输入的验证码。
// 它会验证验证码是否正确，如果正确，则将密钥与用户 ID 关联并存储到数据库。
//
// 安全检查:
// 1. Request Secret Verification: 验证请求是否来自可信源 (内部服务)。
// 2. Content-Type Header Verification (JSON): 确保请求体是 JSON 格式。
// 3. User Existence Check: 确保要注册 TOTP 的用户存在。
// 4. Key Format & Length Check: 验证提供的密钥是否是有效的 Base64 编码，且解码后长度符合预期 (通常是 20 字节)。
// 5. Code Presence Check: 确保用户提供了验证码。
// 6. TOTP Code Verification: 使用提供的密钥验证用户输入的验证码是否在允许的时间窗口内有效。
//
// 参数:
//   env (*Environment): 应用环境，包含数据库连接、配置等。
//   w (http.ResponseWriter): HTTP 响应写入器。
//   r (*http.Request): 收到的 HTTP 请求。
//   params (httprouter.Params): URL 参数，包含 'user_id'。
func handleRegisterTOTPRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// 1. 验证内部请求密钥
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	// 2. 验证 Content-Type
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}

	// 从 URL 获取用户 ID
	userId := params.ByName("user_id")
	// 3. 检查用户是否存在
	userExists, err := checkUserExists(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !userExists {
		writeNotFoundErrorResponse(w)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	// 定义解析 JSON 的结构体
	var data struct {
		Key  *string `json:"key"`  // Base64 编码的 TOTP 密钥
		Code *string `json:"code"` // 用户输入的当前 TOTP 验证码
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	// 检查密钥是否存在
	if data.Key == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	// 4. 解码 Base64 密钥
	key, err := base64.StdEncoding.DecodeString(*data.Key)
	if err != nil {
		// Base64 解码失败，说明密钥格式无效
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	// 检查解码后的密钥长度是否为 20 字节 (常见的 TOTP 密钥长度)
	if len(key) != 20 {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	// 5. 检查验证码是否存在且不为空
	if data.Code == nil || *data.Code == "" {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	// 6. 验证 TOTP 验证码
	// 使用 otp 包验证，允许前后 10 秒的容错时间窗口 (grace period)
	validCode := otp.VerifyTOTPWithGracePeriod(time.Now(), key, 30*time.Second, 6, *data.Code, 10*time.Second)
	if !validCode {
		// 验证码不正确
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectCode)
		return
	}

	// 验证码正确，将密钥注册到数据库
	credential, err := registerUserTOTPCredential(env.db, r.Context(), userId, key)
	if errors.Is(err, ErrRecordNotFound) {
		// 这个错误理论上不应该在这里发生，因为前面已经检查过 userExists
		// 但以防万一，如果 register 函数内部再次检查并发现用户不存在，则返回 404
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		// 其他数据库错误
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// 注册成功，返回包含凭据信息的 JSON (通常只包含 ID 和创建时间，不含密钥)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(credential.EncodeToJSON()))
}

// handleVerifyTOTPRequest 处理用户登录时验证 TOTP 验证码的 API 请求。
// 当用户启用了 2FA 并已成功输入密码后，需要再输入当前的 TOTP 验证码进行验证。
// 此函数接收用户 ID 和用户输入的验证码。
// 它会从数据库获取该用户的 TOTP 密钥，然后使用密钥验证用户输入的验证码。
//
// 安全检查:
// 1. Request Secret Verification.
// 2. Content-Type Header Verification (JSON).
// 3. User Existence Check.
// 4. TOTP Credential Existence Check: 检查用户是否已注册 TOTP。
// 5. Code Presence Check.
// 6. Rate Limiting (per User): 限制单个用户尝试验证 TOTP 的频率，防止暴力猜测。
// 7. TOTP Code Verification: 使用存储的密钥验证用户输入的验证码。
//
// 参数:
//   env (*Environment): 应用环境。
//   w (http.ResponseWriter): HTTP 响应写入器。
//   r (*http.Request): 收到的 HTTP 请求。
//   params (httprouter.Params): URL 参数，包含 'user_id'。
func handleVerifyTOTPRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// 1. 验证内部请求密钥
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	// 2. 验证 Content-Type
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}

	// 从 URL 获取用户 ID
	userId := params.ByName("user_id")
	// 3. 检查用户是否存在
	userExists, err := checkUserExists(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !userExists {
		writeNotFoundErrorResponse(w)
		return
	}

	// 4. 获取用户的 TOTP 凭据 (包含密钥)
	credential, err := getUserTOTPCredential(env.db, r.Context(), userId)
	if errors.Is(err, ErrRecordNotFound) {
		// 如果用户没有注册 TOTP，返回不允许操作 (或特定的错误码表明未设置 2FA)
		writeExpectedErrorResponse(w, ExpectedErrorNotAllowed)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	// 定义解析 JSON 的结构体
	var data struct {
		Code *string `json:"code"` // 用户输入的当前 TOTP 验证码
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	// 5. 检查验证码是否存在且不为空
	if data.Code == nil || *data.Code == "" {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	// 6. 应用针对用户的速率限制
	if !env.totpUserRateLimit.Consume(userId) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	// 7. 验证 TOTP 验证码
	valid := otp.VerifyTOTPWithGracePeriod(time.Now(), credential.Key, 30*time.Second, 6, *data.Code, 10*time.Second)
	if !valid {
		// 验证码不正确
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectCode)
		return
	}
	// 验证成功，重置该用户的速率限制计数器
	env.totpUserRateLimit.Reset(userId)

	// 验证成功，返回 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// handleDeleteUserTOTPCredentialRequest 处理删除用户 TOTP 凭据的 API 请求。
// 用户可能希望禁用 2FA，这时需要删除存储的 TOTP 密钥。
//
// 安全检查:
// 1. Request Secret Verification.
// 2. TOTP Credential Existence Check: 确保用户确实设置了 TOTP 才能删除。
//
// 参数:
//   env (*Environment): 应用环境。
//   w (http.ResponseWriter): HTTP 响应写入器。
//   r (*http.Request): 收到的 HTTP 请求。
//   params (httprouter.Params): URL 参数，包含 'user_id'。
func handleDeleteUserTOTPCredentialRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// 1. 验证内部请求密钥
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}

	// 从 URL 获取用户 ID
	userId := params.ByName("user_id")
	// 2. 检查用户的 TOTP 凭据是否存在
	_, err := getUserTOTPCredential(env.db, r.Context(), userId)
	if errors.Is(err, ErrRecordNotFound) {
		// 如果凭据本就不存在，返回 404 Not Found
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// 凭据存在，执行删除操作
	err = deleteUserTOTPCredential(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// 删除成功，返回 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// handleGetUserTOTPCredentialRequest 处理获取用户 TOTP 凭据信息的 API 请求。
// 注意：此接口通常只应返回非敏感信息，如凭据是否存在、创建时间等，**绝不能返回密钥本身**。
// 它的主要用途可能是让客户端检查用户是否已启用 TOTP。
//
// 安全检查:
// 1. Request Secret Verification.
// 2. Accept Header Verification (JSON): 期望客户端接受 JSON 响应。
// 3. TOTP Credential Existence Check.
//
// 参数:
//   env (*Environment): 应用环境。
//   w (http.ResponseWriter): HTTP 响应写入器。
//   r (*http.Request): 收到的 HTTP 请求。
//   params (httprouter.Params): URL 参数，包含 'user_id'。
func handleGetUserTOTPCredentialRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// 1. 验证内部请求密钥
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	// 2. 验证 Accept 头
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}
	// 从 URL 获取用户 ID
	userId := params.ByName("user_id")
	// 3. 获取用户的 TOTP 凭据
	credential, err := getUserTOTPCredential(env.db, r.Context(), userId)
	if errors.Is(err, ErrRecordNotFound) {
		// 如果凭据不存在，返回 404 Not Found
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// 凭据存在，返回编码后的 JSON 信息 (不含密钥)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(credential.EncodeToJSON()))
}

// --- 数据库操作函数 ---

// getUserTOTPCredential 根据用户 ID 从数据库中检索用户的 TOTP 凭据。
//
// 参数:
//   db (*sql.DB): 数据库连接池。
//   ctx (context.Context): 请求上下文。
//   userId (string): 要检索凭据的用户 ID。
//
// 返回值:
//   UserTOTPCredential: 找到的用户 TOTP 凭据对象。
//   error: 如果查询时发生错误或未找到记录 (ErrRecordNotFound)，则返回错误。
func getUserTOTPCredential(db *sql.DB, ctx context.Context, userId string) (UserTOTPCredential, error) {
	var credential UserTOTPCredential
	var createdAt int64
	// 查询 user_totp_credential 表
	err := db.QueryRowContext(ctx, "SELECT user_id, created_at, key FROM user_totp_credential WHERE user_id = ?", userId).Scan(&credential.UserId, &createdAt, &credential.Key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserTOTPCredential{}, ErrRecordNotFound
		}
		return UserTOTPCredential{}, err
	}
	// 转换时间戳
	credential.CreatedAt = time.Unix(createdAt, 0)
	return credential, nil
}

// registerUserTOTPCredential 在数据库中为用户注册（插入）一个新的 TOTP 凭据。
// 如果用户已存在 TOTP 凭据，此操作可能会失败（取决于数据库约束，通常 user_id 是主键或唯一键）。
//
// 参数:
//   db (*sql.DB): 数据库连接池。
//   ctx (context.Context): 请求上下文。
//   userId (string): 要注册凭据的用户 ID。
//   key ([]byte): TOTP 密钥（原始字节）。
//
// 返回值:
//   UserTOTPCredential: 创建成功的凭据对象。
//   error: 如果插入数据库时发生错误（如违反唯一约束），则返回错误。
func registerUserTOTPCredential(db *sql.DB, ctx context.Context, userId string, key []byte) (UserTOTPCredential, error) {
	now := time.Now()
	credential := UserTOTPCredential{
		UserId:    userId,
		CreatedAt: now,
		Key:       key, // 直接存储原始密钥字节
	}
	// 插入数据库
	_, err := db.ExecContext(ctx, "INSERT INTO user_totp_credential (user_id, created_at, key) VALUES (?, ?, ?)", credential.UserId, credential.CreatedAt.Unix(), credential.Key)
	if err != nil {
		return UserTOTPCredential{}, err
	}
	return credential, nil
}

// deleteUserTOTPCredential 根据用户 ID 从数据库中删除用户的 TOTP 凭据。
//
// 参数:
//   db (*sql.DB): 数据库连接池。
//   ctx (context.Context): 请求上下文。
//   userId (string): 要删除凭据的用户 ID。
//
// 返回值:
//   error: 如果执行 SQL 删除语句时发生错误，则返回错误。
func deleteUserTOTPCredential(db *sql.DB, ctx context.Context, userId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM user_totp_credential WHERE user_id = ?", userId)
	return err
}

// UserTOTPCredential 定义了存储在数据库中的用户 TOTP 凭据结构。
type UserTOTPCredential struct {
	UserId    string    `json:"user_id"`    // 关联的用户 ID
	CreatedAt time.Time `json:"created_at"` // 凭据创建时间
	Key       []byte    `json:"-"`         // TOTP 密钥 (原始字节), JSON 序列化时忽略此字段 (`json:"-"`) 以防泄露
}

// EncodeToJSON 将 UserTOTPCredential 对象序列化为 JSON 字符串。
// 注意：它显式地忽略了 Key 字段，确保密钥不会包含在 API 响应中。
func (c *UserTOTPCredential) EncodeToJSON() string {
	// 创建一个临时结构体，只包含需要暴露的字段
	data := struct {
		UserId    string `json:"user_id"`
		CreatedAt int64  `json:"created_at"` // 返回 Unix 时间戳
	}{
		UserId:    c.UserId,
		CreatedAt: c.CreatedAt.Unix(),
	}
	// 编码为 JSON
	encoded, err := json.Marshal(data)
	if err != nil {
		// 理论上这个简单的结构体编码不应失败，但以防万一
		return "{}" // 返回空 JSON 对象
	}
	return string(encoded)
}
