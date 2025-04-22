{{ ... }}

// CreateApp initializes the application's main router and registers all API endpoints.
// It uses the custom `Router` wrapper to ensure the `Environment` is available to handlers.
// Each `Handle` call maps an HTTP method and path pattern to a specific handler function (defined elsewhere, likely in auth.go or similar).

// CreateApp 函数负责初始化整个 Faroe 应用的核心路由。
// 想象一下，这就像是应用程序的“总调度室”，它告诉服务器收到什么样的网络请求（比如用户想注册、登录或重置密码）时，
// 应该调用哪个“处理部门”（也就是具体的 handler 函数）去干活。
//
// 这个函数的作用非常关键，它定义了所有对外提供的 API 接口，决定了 Faroe 能做什么。
//
// 参数:
//   env *Environment: 这是一个包含应用运行所需配置和资源的结构体，比如数据库连接、密钥、邮件发送设置等。
//                   所有处理具体请求的 handler 函数都能访问到这个环境信息。
//
// 返回值:
//   http.Handler: 这是 Go 语言里标准的处理 HTTP 请求的接口类型。返回的这个 handler 可以被 Go 的标准
//                 `http.ListenAndServe` 函数用来启动一个 Web 服务器，监听来自客户端（比如你的网站前端或手机 App）的请求。
//
// 工作流程:
// 1. 初始化一个自定义的 Router: 我们没有直接用 Go 标准的路由，而是用了一个叫 `NewRouter` 的东西。
//    这个自定义 Router 的好处是它能把 `Environment` 自动传递给每个请求处理函数，省去了手动传递的麻烦。
//    它还设置了一个“默认处理程序”，当收到的请求路径没有匹配到下面任何一个具体的 API 规则时，就会执行这个默认处理。
//    这里的默认处理是返回一个 404 Not Found 错误，告诉客户端请求的地址不存在。
//    (注释掉的代码示例展示了如何在这里加入一个安全检查，比如验证请求是否带有正确的密钥)。
// 2. 注册各个 API 端点 (Endpoints): 使用 `router.Handle` 方法，把 HTTP 请求方法 (GET, POST, DELETE 等)、
//    URL 路径 (比如 "/users", "/users/:user_id/verify-password") 和对应的处理函数 (比如 handleCreateUserRequest) 关联起来。
//    每个 Handle 调用都定义了一个 Faroe 能响应的具体操作。
//    - `:user_id`, `:request_id` 这种是路径参数，意味着客户端请求时需要在这里填入具体的用户 ID 或请求 ID。
//    - 每个路径后面跟着的处理函数名 (e.g., handleCreateUserRequest) 实际上是在其他 Go 文件 (如 user.go, auth.go 等) 中定义的，
//      这里只是把它们“挂载”到对应的 URL 上。
// 3. 返回配置好的 Handler: 最后，`router.Handler()` 方法会生成一个标准的 http.Handler，包含了所有注册好的路由规则。
func CreateApp(env *Environment) http.Handler {
	// 初始化自定义路由，传入环境配置和默认处理函数
	router := NewRouter(env, func(env *Environment, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// 这个是默认的处理函数，当没有其他路由规则匹配时会执行
		// 这里的示例是直接返回 404 Not Found 错误
		// 实际应用中，这里可能还会做一些基础的请求验证
		// // 比如检查请求是否携带了正确的 API 密钥
		// if !verifyRequestSecret(env.secret, r) {
		// 	writeNotAuthenticatedErrorResponse(w) // 写入未授权错误
		// 	return
		// }
		writeNotFoundErrorResponse(w) // 写入 404 Not Found 错误
	})

	// --- 公共/根路径端点 ---
	// GET /: 这是最基础的访问路径。通常用来做个简单的“健康检查”，看看服务是不是还活着，
	// 或者返回一些基本信息，比如版本号。
	// 这里直接返回 Faroe 的版本号和一个文档链接。
	router.Handle("GET", "/", func(env *Environment, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// // 实际可能需要验证访问密钥
		// if !verifyRequestSecret(env.secret, r) {
		// 	writeNotAuthenticatedErrorResponse(w)
		//  return
		// }
		// 向响应体写入版本信息和文档链接
		w.Write([]byte(fmt.Sprintf("Faroe version %s\nRead the documentation: https://faroe.dev\n", version)))
	})

	// --- 用户管理相关的 API 端点 ---
	// 这些接口用来管理 Faroe 里的用户账号

	// POST /users: 创建一个新用户账号。
	// 客户端需要发送 POST 请求到 /users 路径，请求体里通常包含邮箱、密码等注册信息。
	// 由 handleCreateUserRequest 函数处理（定义在别处）。
	router.Handle("POST", "/users", handleCreateUserRequest)

	// GET /users: 获取用户列表。
	// 这个接口可能需要管理员权限或特殊的访问密钥才能调用。
	// 由 handleGetUsersRequest 函数处理。
	router.Handle("GET", "/users", handleGetUsersRequest)

	// DELETE /users: 批量删除用户。
	// 同样，通常需要管理员权限。
	// 由 handleDeleteUsersRequest 函数处理。
	router.Handle("DELETE", "/users", handleDeleteUsersRequest)

	// GET /users/:user_id: 获取指定 ID 用户的信息。
	// `:user_id` 是一个占位符，请求时需要替换成实际的用户 ID，比如 /users/123。
	// 由 handleGetUserRequest 函数处理。
	router.Handle("GET", "/users/:user_id", handleGetUserRequest)

	// DELETE /users/:user_id: 删除指定 ID 的用户。
	// 由 handleDeleteUserRequest 函数处理。
	router.Handle("DELETE", "/users/:user_id", handleDeleteUserRequest)

	// --- 认证和密码管理相关的 API 端点 ---
	// 这些接口处理用户的登录验证、密码修改、密码重置等功能

	// POST /users/:user_id/verify-password: 验证用户当前密码是否正确。
	// 比如在修改敏感信息前，可能需要用户再输一次密码确认身份。
	// 由 handleVerifyUserPasswordRequest 函数处理。
	router.Handle("POST", "/users/:user_id/verify-password", handleVerifyUserPasswordRequest)

	// POST /users/:user_id/update-password: 更新用户的密码。
	// 可能需要提供旧密码，或者一个有效的密码重置凭证。
	// 由 handleUpdateUserPasswordRequest 函数处理。
	router.Handle("POST", "/users/:user_id/update-password", handleUpdateUserPasswordRequest)

	// POST /users/:user_id/password-reset-requests: 为指定用户发起一个密码重置请求。
	// 这通常会触发发送一封包含重置链接或验证码的邮件给用户。
	// 由 handleCreateUserPasswordResetRequestRequest 函数处理。
	router.Handle("POST", "/users/:user_id/password-reset-requests", handleCreateUserPasswordResetRequestRequest)

	// GET /users/:user_id/password-reset-requests: 查询指定用户的密码重置请求记录。
	// 由 handleGetUserPasswordResetRequestsRequest 函数处理。
	router.Handle("GET", "/users/:user_id/password-reset-requests", handleGetUserPasswordResetRequestsRequest)

	// DELETE /users/:user_id/password-reset-requests: 删除指定用户的密码重置请求记录。
	// 比如用户取消了重置，或者请求已过期。
	// 由 handleDeleteUserPasswordResetRequestsRequest 函数处理。
	router.Handle("DELETE", "/users/:user_id/password-reset-requests", handleDeleteUserPasswordResetRequestsRequest)

	// GET /password-reset-requests/:request_id: 获取某个具体的密码重置请求的详细信息。
	// `:request_id` 是密码重置请求的唯一标识。
	// 由 handleGetPasswordResetRequestRequest 函数处理。
	router.Handle("GET", "/password-reset-requests/:request_id", handleGetPasswordResetRequestRequest)

	// DELETE /password-reset-requests/:request_id: 删除（或作废）一个具体的密码重置请求。
	// 由 handleDeletePasswordResetRequestRequest 函数处理。
	router.Handle("DELETE", "/password-reset-requests/:request_id", handleDeletePasswordResetRequestRequest)

	// POST /password-reset-requests/:request_id/verify-email: 验证与密码重置请求关联的邮箱。
	// 这通常是密码重置流程中的一步，用户点击邮件里的链接会访问这个接口。
	// 由 handleVerifyPasswordResetRequestEmailRequest 函数处理。
	router.Handle("POST", "/password-reset-requests/:request_id/verify-email", handleVerifyPasswordResetRequestEmailRequest)

	// POST /reset-password: 使用一个有效的密码重置凭证（比如验证码或 token）来设置新密码。
	// 这是密码重置流程的最后一步。
	// 由 handleResetPasswordRequest 函数处理。
	router.Handle("POST", "/reset-password", handleResetPasswordRequest)

	// --- 两步验证 (2FA) 相关的 API 端点 ---
	// 这些接口处理基于时间的一次性密码 (TOTP) 的注册、验证和管理

	// POST /users/:user_id/register-totp: 为用户注册一个新的 TOTP 设备（比如手机上的 Authenticator App）。
	// 这个过程通常会生成一个二维码或密钥让用户扫描/输入。
	// 由 handleRegisterTOTPRequest 函数处理。
	router.Handle("POST", "/users/:user_id/register-totp", handleRegisterTOTPRequest)

	// GET /users/:user_id/totp-credential: 获取用户已注册的 TOTP 凭证信息。
	// 比如用来在设置页面显示“两步验证已启用”。
	// 由 handleGetUserTOTPCredentialRequest 函数处理。
	router.Handle("GET", "/users/:user_id/totp-credential", handleGetUserTOTPCredentialRequest)

	// DELETE /users/:user_id/totp-credential: 移除用户的 TOTP 凭证（禁用两步验证）。
	// 由 handleDeleteUserTOTPCredentialRequest 函数处理。
	router.Handle("DELETE", "/users/:user_id/totp-credential", handleDeleteUserTOTPCredentialRequest)

	// POST /users/:user_id/verify-2fa/totp: 验证用户输入的 TOTP 动态验证码是否正确。
	// 在登录或其他需要增强安全性的操作时使用。
	// 由 handleVerifyTOTPRequest 函数处理。
	router.Handle("POST", "/users/:user_id/verify-2fa/totp", handleVerifyTOTPRequest)

	// POST /users/:user_id/reset-2fa: 重置用户的两步验证设置。
	// 可能是管理员操作，或者是用户通过备用码等方式发起的恢复流程。
	// 由 handleResetUser2FARequest 函数处理。
	router.Handle("POST", "/users/:user_id/reset-2fa", handleResetUser2FARequest)

	// POST /users/:user_id/regenerate-recovery-code: 为用户生成新的备用恢复码。
	// 当用户丢失了 TOTP 设备时，可以用恢复码登录并重置 2FA。
	// 由 handleRegenerateUserRecoveryCodeRequest 函数处理。
	router.Handle("POST", "/users/:user_id/regenerate-recovery-code", handleRegenerateUserRecoveryCodeRequest)

	// --- 邮箱验证和更新相关的 API 端点 ---
	// 这些接口处理用户注册邮箱的验证，以及后续修改邮箱地址的流程

	// POST /users/:user_id/email-verification-request: 为用户当前的注册邮箱发起一个验证请求。
	// 通常是新用户注册后，或邮箱状态变为未验证时使用。会发送验证邮件。
	// 由 handleCreateUserEmailVerificationRequestRequest 函数处理。
	router.Handle("POST", "/users/:user_id/email-verification-request", handleCreateUserEmailVerificationRequestRequest)

	// GET /users/:user_id/email-verification-request: 查询用户的邮箱验证请求状态。
	// 由 handleGetUserEmailVerificationRequestRequest 函数处理。
	router.Handle("GET", "/users/:user_id/email-verification-request", handleGetUserEmailVerificationRequestRequest)

	// DELETE /users/:user_id/email-verification-request: 取消或删除用户的邮箱验证请求。
	// 由 handleDeleteUserEmailVerificationRequestRequest 函数处理。
	router.Handle("DELETE", "/users/:user_id/email-verification-request", handleDeleteUserEmailVerificationRequestRequest)

	// POST /users/:user_id/verify-email: 使用发送到用户邮箱的验证码或 token 来完成邮箱验证。
	// 用户点击邮件中的链接或输入验证码时会调用此接口。
	// 由 handleVerifyUserEmailRequest 函数处理。
	router.Handle("POST", "/users/:user_id/verify-email", handleVerifyUserEmailRequest)

	// POST /users/:user_id/email-update-requests: 发起一个更改用户注册邮箱的请求。
	// 通常需要提供新的邮箱地址，并可能需要验证旧邮箱或密码。会向新邮箱发送验证邮件。
	// 由 handleCreateUserEmailUpdateRequestRequest 函数处理。
	router.Handle("POST", "/users/:user_id/email-update-requests", handleCreateUserEmailUpdateRequestRequest)

	// GET /users/:user_id/email-update-requests: 查询用户发起的邮箱更改请求的状态。
	// 由 handleGetUserEmailUpdateRequestsRequest 函数处理。
	router.Handle("GET", "/users/:user_id/email-update-requests", handleGetUserEmailUpdateRequestsRequest)

	// DELETE /users/:user_id/email-update-requests: 取消或删除用户的邮箱更改请求。
	// 由 handleDeleteUserEmailUpdateRequestsRequest 函数处理。
	router.Handle("DELETE", "/users/:user_id/email-update-requests", handleDeleteUserEmailUpdateRequestsRequest)

	// GET /email-update-requests/:request_id: 获取某个具体的邮箱更改请求的详细信息。
	// `:request_id` 是邮箱更改请求的唯一标识。
	// 由 handleGetEmailUpdateRequestRequest 函数处理。
	router.Handle("GET", "/email-update-requests/:request_id", handleGetEmailUpdateRequestRequest)

	// DELETE /email-update-requests/:request_id: 取消或删除一个具体的邮箱更改请求。
	// 由 handleDeleteEmailUpdateRequestRequest 函数处理。
	router.Handle("DELETE", "/email-update-requests/:request_id", handleDeleteEmailUpdateRequestRequest)

	// POST /verify-new-email: 使用发送到 *新* 邮箱的验证码或 token 来完成邮箱地址的更改。
	// 这是邮箱更改流程的最后一步，确认新邮箱有效并完成更新。
	// 由 handleUpdateEmailRequest 函数处理。
	router.Handle("POST", "/verify-new-email", handleUpdateEmailRequest)


	// 所有路由规则都注册完毕后，调用 router.Handler() 生成最终的 http.Handler 并返回。
	// 这个返回的 Handler 就可以交给 Go 的 HTTP 服务器去运行了。
	return router.Handler()
}
