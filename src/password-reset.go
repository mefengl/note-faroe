package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"faroe/argon2id"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter" // 高性能的 HTTP 请求路由器
)

// handleCreateUserPasswordResetRequestRequest 处理创建用户密码重置请求的 API 调用。
// 它首先验证请求的合法性，然后为用户生成一个安全的重置代码，并将代码的哈希值存储到数据库中，
// 最后将包含原始代码（用于发送给用户）和请求详情的 JSON 返回给调用者。
//
// 安全检查:
// 1. Request Secret Verification: 验证请求头中的共享密钥。
// 2. Content-Type & Accept Header Verification: 确保是 JSON 请求和响应。
// 3. User Existence Check: 验证目标用户是否存在。
// 4. Rate Limiting (可选, 基于 ClientIP):
//    - 限制密码哈希相关的操作频率 (passwordHashingIPRateLimit)。
//    - 限制创建密码重置请求的频率 (createPasswordResetIPRateLimit)。
// 5. Expired Request Cleanup: 在创建新请求前，删除该用户已过期的旧请求。
// 6. Secure Code Generation: 使用 crypto/rand 生成安全的验证码。
// 7. Code Hashing: 使用 Argon2id 对验证码进行哈希，只存储哈希值，不存储明文验证码。
//
// 参数:
//   env (*Environment): 应用环境，包含数据库连接、密钥、速率限制器等。
//   w (http.ResponseWriter): HTTP 响应写入器。
//   r (*http.Request): 收到的 HTTP 请求。
//   params (httprouter.Params): URL 参数，包含 'user_id'。
func handleCreateUserPasswordResetRequestRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// 1. 验证请求密钥
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	// 2. 验证 Content-Type
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}
	// 3. 验证 Accept 头
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	// 从 URL 获取用户 ID
	userId := params.ByName("user_id")
	// 4. 检查用户是否存在
	userExists, err := checkUserExists(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !userExists {
		writeNotFoundErrorResponse(w) // 用户不存在，返回 404
		return
	}

	// 尝试读取请求体，以获取可选的 client_ip 用于速率限制
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// 读取请求体失败，通常是无效数据
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	// 5. 如果请求体不为空，尝试解析 client_ip 并应用速率限制
	if len(body) > 0 {
		var data struct {
			ClientIP string `json:"client_ip"` // 从 JSON 中获取客户端 IP
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			// JSON 解析失败
			writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
			return
		}

		// 如果提供了 ClientIP，则进行速率限制检查
		if data.ClientIP != "" {
			// 检查密码哈希相关的速率限制
			if !env.passwordHashingIPRateLimit.Consume(data.ClientIP) {
				writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
				return
			}
			// 检查创建密码重置请求的速率限制
			if !env.createPasswordResetIPRateLimit.Consume(data.ClientIP) {
				writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
				return
			}
		}
	}

	// 6. 删除该用户已过期的密码重置请求
	err = deleteExpiredUserPasswordResetRequests(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// 7. 生成一个安全、随机的验证码
	code, err := generateSecureCode()
	if err != nil {
		log.Println(err) // 记录生成验证码时的错误
		writeUnexpectedErrorResponse(w)
		return
	}

	// 8. 使用 Argon2id 对验证码进行哈希处理
	codeHash, err := argon2id.Hash(code)
	if err != nil {
		log.Println(err) // 记录哈希处理时的错误
		writeUnexpectedErrorResponse(w)
		return
	}

	// 9. 在数据库中创建密码重置请求记录，存储用户ID和验证码哈希
	resetRequest, err := createPasswordResetRequest(env.db, r.Context(), userId, codeHash)
	if err != nil {
		log.Println(err) // 记录数据库插入错误
		writeUnexpectedErrorResponse(w)
		return
	}

	// 10. 成功响应：返回状态码 200 和包含请求详情及 *原始验证码* 的 JSON
	// 注意：这里返回原始验证码 code 是为了让调用方（例如后端服务）能够将其发送给用户（通过邮件等方式）
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 使用常量 http.StatusOK 更清晰
	w.Write([]byte(resetRequest.EncodeToJSONWithCode(code))) // 使用带 code 的编码方法
}

// handleGetPasswordResetRequestRequest 处理获取特定密码重置请求详情的 API 调用。
// 它根据请求 ID 查找记录，并检查是否过期。
//
// 安全检查:
// 1. Request Secret Verification.
// 2. Accept Header Verification (JSON).
// 3. Request Existence Check.
// 4. Expiry Check: 如果请求已过期，则将其删除并返回 404。
//
// 参数:
//   env (*Environment): 应用环境。
//   w (http.ResponseWriter): HTTP 响应写入器。
//   r (*http.Request): 收到的 HTTP 请求。
//   params (httprouter.Params): URL 参数，包含 'request_id'。
func handleGetPasswordResetRequestRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// 1. 验证请求密钥
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	// 2. 验证 Accept 头
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	// 从 URL 获取请求 ID
	resetRequestId := params.ByName("request_id")
	// 3. 从数据库获取密码重置请求
	resetRequest, err := getPasswordResetRequest(env.db, r.Context(), resetRequestId)
	if errors.Is(err, ErrRecordNotFound) {
		// 请求未找到
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		// 其他数据库错误
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	// 4. 检查请求是否已过期
	// time.Now().Compare(t) 返回: -1 (now < t), 0 (now == t), 1 (now > t)
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 { // 如果当前时间晚于或等于过期时间
		// 尝试删除已过期的请求
		err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
		if err != nil {
			// 记录删除错误，但仍然按过期处理
			log.Println(err)
			// 注意：这里原代码返回了 UnexpectedError，但逻辑上应该返回 404，因为请求已失效
			// writeUnexpectedErrorResponse(w)
			// return
		}
		// 返回 404 Not Found，表示请求无效（已过期）
		writeNotFoundErrorResponse(w)
		return
	}
	// 5. 成功响应：返回请求详情（不包含验证码）
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	w.Write([]byte(resetRequest.EncodeToJSON()))
}

// handleVerifyPasswordResetRequestEmailRequest 处理验证密码重置代码的 API 调用。
// 用户提供请求 ID 和他们收到的验证码，此函数验证代码是否与数据库中存储的哈希匹配，并检查请求是否过期。
// 它还应用了针对单个重置请求 ID 的尝试次数限制。
//
// 安全检查:
// 1. Request Secret Verification.
// 2. Content-Type Header Verification (JSON).
// 3. Request Existence Check.
// 4. Expiry Check.
// 5. Code Presence Check: 确保请求体中包含 'code'。
// 6. Rate Limiting (可选, 基于 ClientIP): 限制密码哈希相关的操作频率。
// 7. Attempt Limiting: 限制对 *同一个* 重置请求 ID 的验证尝试次数 (verifyPasswordResetCodeLimitCounter)。
//    如果超过限制，请求将被删除。
// 8. Code Validation: 使用 Argon2id.Verify 对比提供的代码和存储的哈希。
//
// 参数:
//   env (*Environment): 应用环境。
//   w (http.ResponseWriter): HTTP 响应写入器。
//   r (*http.Request): 收到的 HTTP 请求。
//   params (httprouter.Params): URL 参数，包含 'request_id'。
func handleVerifyPasswordResetRequestEmailRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// 1. 验证请求密钥
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	// 2. 验证 Content-Type
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}

	// 从 URL 获取请求 ID
	resetRequestId := params.ByName("request_id")
	// 3. 获取密码重置请求
	resetRequest, err := getPasswordResetRequest(env.db, r.Context(), resetRequestId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	// 4. 检查请求是否已过期
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		// 尝试删除已过期的请求
		err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
		if err != nil {
			log.Println(err)
			// 同样，这里原代码返回 UnexpectedError，改为返回 404 更合理
			// writeUnexpectedErrorResponse(w)
			// return
		}
		writeNotFoundErrorResponse(w)
		return
	}

	// 读取请求体以获取验证码和可选的 ClientIP
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	// 定义用于解析 JSON 的结构体
	var data struct {
		Code     *string `json:"code"`      // 用户提供的验证码 (指针以区分空字符串和未提供)
		ClientIP string  `json:"client_ip"` // 可选的客户端 IP，用于速率限制
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		// JSON 解析失败
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	// 5. 检查验证码是否提供且不为空
	if data.Code == nil || *data.Code == "" {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	// 6. 应用基于 IP 的密码哈希速率限制（如果提供了 IP）
	if data.ClientIP != "" && !env.passwordHashingIPRateLimit.Consume(data.ClientIP) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}

	// 7. 应用基于请求 ID 的验证尝试次数限制
	// consume 方法会减少计数器的值，如果减到 0 以下则返回 false
	if !env.verifyPasswordResetCodeLimitCounter.Consume(resetRequest.Id) {
		// 尝试次数超限，删除此重置请求，使其失效
		err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
		if err != nil {
			// 记录删除错误，但仍然按超限处理
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		// 返回请求过多错误
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}

	// 8. 使用 Argon2id 验证提供的代码是否与存储的哈希匹配
	validCode, err := argon2id.Verify(resetRequest.CodeHash, *data.Code)
	if err != nil {
		// 验证过程中发生内部错误
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// 如果验证码不正确
	if !validCode {
		// 返回密码不正确（这里复用了密码错误，也可以定义专门的验证码错误）
		writeExpectedErrorResponse(w, ExpectedErrorIncorrectPassword)
		return
	}

	// 验证成功！
	// 重置该请求 ID 的尝试次数限制计数器
	env.verifyPasswordResetCodeLimitCounter.AddTokenIfEmpty(resetRequest.Id)

	// 响应 204 No Content，表示验证成功，无需返回内容
	w.WriteHeader(http.StatusNoContent)
}

func handleResetPasswordRequest(env *Environment, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	var data struct {
		RequestId *string `json:"request_id"`
		Password  *string `json:"password"`
		ClientIP  string  `json:"client_ip"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	if data.RequestId == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	if data.Password == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	resetRequest, err := getPasswordResetRequest(env.db, r.Context(), *data.RequestId)
	if errors.Is(err, ErrRecordNotFound) {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidRequest)
		return
	}
	if err != nil {
		writeUnexpectedErrorResponse(w)
		return
	}
	// If now is or after expiration
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeExpectedErrorResponse(w, ExpectedErrorInvalidRequest)
		return
	}

	password := *data.Password
	if len(password) > 127 {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	strongPassword, err := verifyPasswordStrength(password)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !strongPassword {
		writeExpectedErrorResponse(w, ExpectedErrorWeakPassword)
		return
	}

	if data.ClientIP != "" && !env.passwordHashingIPRateLimit.Consume(data.ClientIP) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}
	passwordHash, err := argon2id.Hash(password)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	validResetRequest, err := resetUserPasswordWithPasswordResetRequest(env.db, r.Context(), resetRequest.Id, passwordHash)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !validResetRequest {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidRequest)
		return
	}

	w.WriteHeader(204)
}

// handleResetPasswordRequest 处理实际重置密码的 API 调用。
// 这个请求通常是在用户成功验证了密码重置代码之后发起的。
// 它需要提供重置请求 ID 和新密码。函数会验证新密码强度，哈希新密码，
// 然后使用重置请求 ID 更新数据库中对应用户的密码哈希，并删除该重置请求。
//
// 注意：这个接口的设计似乎有点问题。
// 它只接收 Request ID 和新密码，但没有验证这个 Request ID 是否真的刚刚被验证通过。
// 更好的做法可能是：
// 1. handleVerifyPasswordResetRequestEmailRequest 验证成功后，返回一个临时的、一次性的令牌。
// 2. handleResetPasswordRequest 需要提供这个一次性令牌和新密码，而不是 Request ID。
// 3. 或者，handleVerifyPasswordResetRequestEmailRequest 验证成功后，直接在这个函数里更新密码，
//    而不是分两步。当前实现可能存在安全风险，即攻击者可以尝试用旧的、但未过期的 Request ID 来重置密码，
//    只要他们能猜到或获取到 Request ID。
//    不过，由于 Request ID 是 UUID，猜到的可能性极低。
//    同时，验证接口 (handleVerify) 做了尝试次数限制，重置接口本身也应该做类似的限制或依赖验证接口的状态。
//    目前的实现看起来依赖于客户端在验证成功后 *立即* 调用重置接口。
//
// 安全检查:
// 1. Request Secret Verification.
// 2. Content-Type Header Verification (JSON).
// 3. Request Existence Check (根据 Request ID)。
// 4. Expiry Check (再次检查，以防万一)。
// 5. New Password Presence & Constraint Check.
// 6. New Password Strength Check.
// 7. Rate Limiting (可选, 基于 ClientIP): 限制密码哈希操作。
// 8. Reset Execution: 使用 `resetUserPasswordWithPasswordResetRequest` 原子地更新密码并删除请求。
//
// 参数:
//   env (*Environment): 应用环境。
//   w (http.ResponseWriter): HTTP 响应写入器。
//   r (*http.Request): 收到的 HTTP 请求。
//   _ (httprouter.Params): URL 参数 (未使用)。
func handleResetPasswordRequest(env *Environment, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// 1. 验证请求密钥
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	// 2. 验证 Content-Type
	if !verifyJSONContentTypeHeader(r) {
		writeUnsupportedMediaTypeErrorResponse(w)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	// 定义解析 JSON 的结构体
	var data struct {
		RequestId    *string `json:"request_id"` // 密码重置请求的 ID
		Password     *string `json:"password"`   // 用户设置的新密码
		ClientIP     string  `json:"client_ip"` // 可选的客户端 IP
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}
	// 检查必需的字段是否提供
	if data.RequestId == nil || *data.RequestId == "" || data.Password == nil {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	// 3. 再次获取密码重置请求，确保它仍然存在且有效
	resetRequest, err := getPasswordResetRequest(env.db, r.Context(), *data.RequestId)
	if errors.Is(err, ErrRecordNotFound) {
		// 如果找不到请求（可能已被删除或过期），返回不允许操作
		writeExpectedErrorResponse(w, ExpectedErrorNotAllowed)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	// 4. 再次检查是否过期
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		// 尝试删除
		err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
		if err != nil {
			log.Println(err)
		}
		// 返回不允许操作
		writeExpectedErrorResponse(w, ExpectedErrorNotAllowed)
		return
	}

	// 5. 检查新密码是否为空或过长
	if *data.Password == "" || len(*data.Password) > 127 {
		writeExpectedErrorResponse(w, ExpectedErrorInvalidData)
		return
	}

	// 6. 检查新密码强度
	strongPassword, err := verifyPasswordStrength(*data.Password)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	if !strongPassword {
		writeExpectedErrorResponse(w, ExpectedErrorWeakPassword)
		return
	}

	// 7. 应用密码哈希的速率限制
	if data.ClientIP != "" && !env.passwordHashingIPRateLimit.Consume(data.ClientIP) {
		writeExpectedErrorResponse(w, ExpectedErrorTooManyRequests)
		return
	}

	// 哈希新密码
	passwordHash, err := argon2id.Hash(*data.Password)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	// 8. 在数据库中执行密码重置操作
	// 这个函数应该原子地更新用户密码并删除重置请求
	ok, err := resetUserPasswordWithPasswordResetRequest(env.db, r.Context(), *data.RequestId, passwordHash)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	// 如果 resetUserPassword... 返回 false，说明重置由于某种原因失败（例如请求已被使用或删除）
	if !ok {
		// 返回不允许操作
		writeExpectedErrorResponse(w, ExpectedErrorNotAllowed)
		return
	}

	// 密码重置成功
	// 响应 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

func handleDeletePasswordResetRequestRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	resetRequestId := params.ByName("request_id")
	resetRequest, err := getPasswordResetRequest(env.db, r.Context(), resetRequestId)
	if errors.Is(err, ErrRecordNotFound) {
		writeNotFoundErrorResponse(w)
		return
	}
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	// If now is or after expiration
	if time.Now().Compare(resetRequest.ExpiresAt) >= 0 {
		err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
		if err != nil {
			log.Println(err)
			writeUnexpectedErrorResponse(w)
			return
		}
		writeNotFoundErrorResponse(w)
		return
	}

	err = deletePasswordResetRequest(env.db, r.Context(), resetRequest.Id)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	w.WriteHeader(204)
}

func handleGetUserPasswordResetRequestsRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
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

	err = deleteExpiredUserPasswordResetRequests(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	resetRequest, err := getUserPasswordResetRequests(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if len(resetRequest) == 0 {
		w.Write([]byte("[]"))
		return
	}
	w.Write([]byte("["))
	for i, user := range resetRequest {
		w.Write([]byte(user.EncodeToJSON()))
		if i != len(resetRequest)-1 {
			w.Write([]byte(","))
		}
	}
	w.Write([]byte("]"))
}

func handleDeleteUserPasswordResetRequestsRequest(env *Environment, w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !verifyRequestSecret(env.secret, r) {
		writeNotAuthenticatedErrorResponse(w)
		return
	}
	if !verifyJSONAcceptHeader(r) {
		writeNotAcceptableErrorResponse(w)
		return
	}

	userId := params.ByName("user_id")
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

	err = deleteUserPasswordResetRequests(env.db, r.Context(), userId)
	if err != nil {
		log.Println(err)
		writeUnexpectedErrorResponse(w)
		return
	}
	w.WriteHeader(204)
}

// createPasswordResetRequest 在数据库中创建一个新的密码重置请求记录。
// 它生成一个唯一的请求 ID (UUID)，设置创建时间和过期时间（通常是当前时间 + 一个固定的有效期），
// 然后调用 insertPasswordResetRequest 将记录插入数据库。
//
// 参数:
//   db (*sql.DB): 数据库连接池。
//   ctx (context.Context): 请求上下文。
//   userId (string): 请求密码重置的用户的 ID。
//   codeHash (string): 使用 Argon2id 哈希过的验证码。
//
// 返回值:
//   PasswordResetRequest: 创建成功的密码重置请求对象。
//   error: 如果生成 UUID 或插入数据库时发生错误，则返回错误。
func createPasswordResetRequest(db *sql.DB, ctx context.Context, userId string, codeHash string) (PasswordResetRequest, error) {
	// 生成一个新的 UUID 作为请求 ID
	requestId, err := newId()
	if err != nil {
		return PasswordResetRequest{}, fmt.Errorf("failed to create password reset request id: %w", err)
	}
	// 获取当前时间
	now := time.Now()
	// 创建 PasswordResetRequest 结构体实例
	request := PasswordResetRequest{
		Id:        requestId,                     // 请求的唯一 ID
		UserId:    userId,                        // 关联的用户 ID
		CreatedAt: now,                         // 创建时间
		ExpiresAt: now.Add(time.Minute * 15), // 过期时间（例如，15分钟后）
		CodeHash:  codeHash,                    // 验证码的 Argon2id 哈希值
	}
	// 将请求记录插入数据库
	err = insertPasswordResetRequest(db, ctx, &request)
	if err != nil {
		return PasswordResetRequest{}, fmt.Errorf("failed to insert password reset request: %w", err)
	}
	// 返回创建的请求对象
	return request, nil
}

// insertPasswordResetRequest 将一个 PasswordResetRequest 对象插入到数据库的 user_password_reset_request 表中。
//
// 参数:
//   db (*sql.DB): 数据库连接池。
//   ctx (context.Context): 请求上下文。
//   request (*PasswordResetRequest): 要插入的密码重置请求对象的指针。
//
// 返回值:
//   error: 如果执行 SQL 插入语句时发生错误，则返回错误。
func insertPasswordResetRequest(db *sql.DB, ctx context.Context, request *PasswordResetRequest) error {
	_, err := db.ExecContext(ctx, "INSERT INTO user_password_reset_request(id, user_id, created_at, expires_at, code_hash) VALUES(?, ?, ?, ?, ?)", request.Id, request.UserId, request.CreatedAt.Unix(), request.ExpiresAt.Unix(), request.CodeHash)
	return err
}

// getPasswordResetRequest 根据请求 ID 从数据库中检索单个密码重置请求记录。
// 如果找不到记录，它会返回 ErrRecordNotFound 错误。
//
// 参数:
//   db (*sql.DB): 数据库连接池。
//   ctx (context.Context): 请求上下文。
//   requestId (string): 要检索的密码重置请求的 ID。
//
// 返回值:
//   PasswordResetRequest: 找到的密码重置请求对象。
//   error: 如果查询时发生错误或未找到记录 (ErrRecordNotFound)，则返回错误。
func getPasswordResetRequest(db *sql.DB, ctx context.Context, requestId string) (PasswordResetRequest, error) {
	var request PasswordResetRequest
	var createdAt int64
	var expiresAt int64
	// 查询数据库
	err := db.QueryRowContext(ctx, "SELECT id, user_id, created_at, expires_at, code_hash FROM user_password_reset_request WHERE id = ?", requestId).Scan(&request.Id, &request.UserId, &createdAt, &expiresAt, &request.CodeHash)
	if err != nil {
		// 如果是没找到记录的错误，返回特定的 ErrRecordNotFound
		if errors.Is(err, sql.ErrNoRows) {
			return PasswordResetRequest{}, ErrRecordNotFound
		}
		// 其他数据库错误
		return PasswordResetRequest{}, err
	}
	// 将 Unix 时间戳转换为 time.Time 对象
	request.CreatedAt = time.Unix(createdAt, 0)
	request.ExpiresAt = time.Unix(expiresAt, 0)
	return request, nil
}

// getUserPasswordResetRequests 根据用户 ID 从数据库中检索该用户的所有未过期的密码重置请求记录。
// 注意：此函数查询的是所有请求，包括已过期的。在 API 层面 (`handleGetUserPasswordResetRequestsRequest`) 通常只返回未过期的，或者这里可以增加 `expires_at > ?` 条件。
// 目前实现是获取所有记录。
//
// 参数:
//   db (*sql.DB): 数据库连接池。
//   ctx (context.Context): 请求上下文。
//   userId (string): 要检索请求的用户 ID。
//
// 返回值:
//   []PasswordResetRequest: 找到的密码重置请求对象切片 (可能为空)。
//   error: 如果查询或扫描数据时发生错误，则返回错误。
func getUserPasswordResetRequests(db *sql.DB, ctx context.Context, userId string) ([]PasswordResetRequest, error) {
	// 查询该用户的所有密码重置请求
	rows, err := db.QueryContext(ctx, "SELECT id, user_id, created_at, expires_at, code_hash FROM user_password_reset_request WHERE user_id = ?", userId)
	if err != nil {
		return nil, err
	}
	// 确保在函数结束时关闭 rows
	defer rows.Close()

	var requests []PasswordResetRequest
	// 遍历查询结果
	for rows.Next() {
		var request PasswordResetRequest
		var createdAt int64
		var expiresAt int64
		// 扫描行数据到结构体
		if err := rows.Scan(&request.Id, &request.UserId, &createdAt, &expiresAt, &request.CodeHash); err != nil {
			// 如果扫描出错，返回错误
			return nil, err
		}
		// 转换时间戳
		request.CreatedAt = time.Unix(createdAt, 0)
		request.ExpiresAt = time.Unix(expiresAt, 0)
		// 将请求添加到切片中
		requests = append(requests, request)
	}
	// 检查遍历过程中是否发生错误
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// 返回找到的请求列表
	return requests, nil
}

func resetUserPasswordWithPasswordResetRequest(db *sql.DB, ctx context.Context, requestId string, passwordHash string) (bool, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	var userId string
	err = tx.QueryRow("DELETE FROM password_reset_request WHERE id = ? AND expires_at > ? RETURNING user_id", requestId, time.Now().Unix()).Scan(&userId)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			return false, err
		}
		return false, nil
	}
	if err != nil {
		tx.Rollback()
		return false, err
	}
	_, err = tx.Exec("DELETE FROM password_reset_request WHERE user_id = ?", userId)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	_, err = tx.Exec("UPDATE user SET password_hash = ? WHERE id = ?", passwordHash, userId)
	if err != nil {
		tx.Rollback()
		return false, err
	}
	tx.Commit()
	return true, nil
}

func deletePasswordResetRequest(db *sql.DB, ctx context.Context, requestId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM password_reset_request WHERE id = ?", requestId)
	return err
}

func deleteExpiredUserPasswordResetRequests(db *sql.DB, ctx context.Context, userId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM password_reset_request WHERE user_id = ? AND expires_at <= ?", userId, time.Now().Unix())
	return err
}

func deleteUserPasswordResetRequests(db *sql.DB, ctx context.Context, userId string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM password_reset_request WHERE user_id = ?", userId)
	return err
}

type PasswordResetRequest struct {
	Id        string
	UserId    string
	CreatedAt time.Time
	ExpiresAt time.Time
	CodeHash  string
}

func (r *PasswordResetRequest) EncodeToJSON() string {
	encoded := fmt.Sprintf("{\"id\":\"%s\",\"user_id\":\"%s\",\"created_at\":%d,\"expires_at\":%d}", r.Id, r.UserId, r.CreatedAt.Unix(), r.ExpiresAt.Unix())
	return encoded
}

func (r *PasswordResetRequest) EncodeToJSONWithCode(code string) string {
	encoded := fmt.Sprintf("{\"id\":\"%s\",\"user_id\":\"%s\",\"created_at\":%d,\"expires_at\":%d,\"code\":\"%s\"}", r.Id, r.UserId, r.CreatedAt.Unix(), r.ExpiresAt.Unix(), code)
	return encoded
}
