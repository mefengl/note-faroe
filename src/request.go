package main

import (
	"crypto/subtle" // 导入用于执行常量时间比较的包，增强安全性
	"mime"          // 导入用于解析 MIME 媒体类型的包
	"net/http"      // 导入处理 HTTP 请求和响应的核心包
	"strings"       // 导入处理字符串操作的包
)

// verifyRequestSecret 函数用于验证 HTTP 请求头中是否包含正确的服务器密钥。
// 这是一种安全措施，确保只有知道密钥的客户端才能访问某些受保护的 API 端点。
// 参数：
//   secret []byte: 服务器配置的密钥，字节切片形式。
//   r *http.Request: 代表客户端发来的 HTTP 请求。
// 返回值：
//   bool: 如果密钥验证通过（或者服务器没有配置密钥），返回 true；否则返回 false。
// 工作原理：
// 1. 检查服务器是否配置了密钥 (len(secret) == 0)。如果没配置，则认为所有请求都合法，直接返回 true。
// 2. 从请求头 (r.Header) 中查找名为 "Authorization" 的字段。
// 3. 如果找不到 "Authorization" 头，或者头的值不是预期的格式，验证失败，返回 false。
// 4. 使用 crypto/subtle.ConstantTimeCompare 进行常量时间比较。这很重要，可以防止"时序攻击" (timing attack)，
//    避免攻击者通过测量比较操作所需的时间来猜测密钥内容。
// 5. 如果比较结果为 1 (表示字节完全匹配)，则验证通过，返回 true；否则返回 false。
func verifyRequestSecret(secret []byte, r *http.Request) bool {
	// 如果服务器没有设置密钥，则认为所有请求都已验证
	if len(secret) == 0 {
		return true
	}
	// 尝试从请求头中获取 "Authorization" 字段的值
	authorizationHeader, ok := r.Header["Authorization"]
	// 如果请求头中没有 "Authorization" 字段，则验证失败
	if !ok {
		return false
	}
	// 使用常量时间比较函数来比较请求头中的值和服务器密钥
	// subtle.ConstantTimeCompare 返回 1 表示相等，0 表示不等
	// 我们只取 Authorization 头的第一个值 (authorizationHeader[0]) 来比较
	return subtle.ConstantTimeCompare(secret, []byte(authorizationHeader[0])) == 1
}

// verifyJSONContentTypeHeader 函数检查 HTTP 请求头中的 "Content-Type" 是否表明
// 请求体的内容是 JSON 格式 (application/json) 或者纯文本 (text/plain)。
// 这有助于服务器正确解析请求体。
// 参数：
//   r *http.Request: 客户端发来的 HTTP 请求。
// 返回值：
//   bool: 如果 Content-Type 是 application/json 或 text/plain，或者请求没有 Content-Type 头，返回 true；
//         如果 Content-Type 无效或不是这两种类型，返回 false。
// 工作原理：
// 1. 尝试获取 "Content-Type" 请求头。
// 2. 如果没有这个头 (ok == false)，默认认为可以通过 (返回 true)。这是因为 GET 等请求可能没有请求体，也就没有 Content-Type。
// 3. 使用 mime.ParseMediaType 解析 Content-Type 头的值。这个函数可以处理像 "application/json; charset=utf-8" 这样的复杂值，
//    提取出主要的媒体类型 (mediatype)，例如 "application/json"。
// 4. 如果解析出错 (err != nil)，说明 Content-Type 格式不正确，返回 false。
// 5. 检查解析出的媒体类型是否是 "application/json" 或 "text/plain"。如果是，返回 true；否则返回 false。
func verifyJSONContentTypeHeader(r *http.Request) bool {
	// 尝试获取 "Content-Type" 请求头
	contentType, ok := r.Header["Content-Type"]
	// 如果没有 Content-Type 头，则默认通过
	if !ok {
		return true
	}
	// 解析 Content-Type 头的值，提取媒体类型部分
	mediatype, _, err := mime.ParseMediaType(contentType[0]) // 只处理第一个 Content-Type 值
	// 如果解析出错，说明格式无效，返回 false
	if err != nil {
		return false
	}
	// 检查媒体类型是否是 "application/json" 或 "text/plain"
	return mediatype == "application/json" || mediatype == "text/plain"
}

// verifyJSONAcceptHeader 函数检查 HTTP 请求头中的 "Accept" 是否表明
// 客户端能够接受 JSON 格式 (application/json) 的响应。
// 服务器可以根据这个头来决定返回什么格式的数据。
// 参数：
//   r *http.Request: 客户端发来的 HTTP 请求。
// 返回值：
//   bool: 如果 Accept 头表明接受 JSON (包括通配符 * / * 或 application/*)，或者请求没有 Accept 头，返回 true；否则返回 false。
// 工作原理：
// 1. 尝试获取 "Accept" 请求头。
// 2. 如果没有 Accept 头 (ok == false)，默认认为客户端能接受任何格式，包括 JSON，返回 true。
// 3. 将 Accept 头的值按逗号 (,) 分割成多个条目 (entries)。一个 Accept 头可能包含多个可接受的类型，例如 "application/json, text/plain, */*"。
// 4. 遍历每个条目：
//    a. 去除条目首尾的空格。
//    b. 按分号 (;) 分割条目，因为 Accept 头可能带有权重因子 (如 application/json;q=0.9)，我们只关心类型本身 (parts[0])。
//    c. 再次去除媒体类型 (mediaType) 首尾的空格。
//    d. 检查媒体类型是否是 "*/*" (接受任何类型), "application/*" (接受任何 application 子类型) 或 "application/json"。
//    e. 如果匹配到任何一个，说明客户端接受 JSON，立即返回 true。
// 5. 如果遍历完所有条目都没有找到匹配的，说明客户端不接受 JSON，返回 false。
func verifyJSONAcceptHeader(r *http.Request) bool {
	// 尝试获取 "Accept" 请求头
	accept, ok := r.Header["Accept"]
	// 如果没有 Accept 头，默认认为客户端接受 JSON
	if !ok {
		return true
	}
	// 按逗号分割 Accept 头的值
	entries := strings.Split(accept[0], ",") // 只处理第一个 Accept 值
	// 遍历每个可接受的媒体类型条目
	for _, entry := range entries {
		// 去除首尾空格
		entry = strings.TrimSpace(entry)
		// 按分号分割，提取媒体类型部分 (忽略可能的权重 q 值)
		parts := strings.Split(entry, ";")
		mediaType := strings.TrimSpace(parts[0])
		// 检查是否接受 JSON (包括通配符)
		if mediaType == "*/*" || mediaType == "application/*" || mediaType == "application/json" {
			// 如果接受，立即返回 true
			return true
		}
	}
	// 如果遍历完所有条目都不接受 JSON，返回 false
	return false
}

// parseJSONOrTextAcceptHeader 函数解析 HTTP 请求头中的 "Accept"，判断客户端
// 是希望接收 JSON 格式还是纯文本 (text/plain) 格式的响应。
// 它优先考虑 JSON。
// 参数：
//   r *http.Request: 客户端发来的 HTTP 请求。
// 返回值：
//   ContentType: 一个整数常量，表示客户端期望的内容类型 (ContentTypeJSON 或 ContentTypePlainText)。
//   bool: 一个布尔值，表示解析是否成功。如果 Accept 头有效且明确指定了 JSON 或 text/plain (或通配符)，返回 true；
//         如果 Accept 头无效或没有明确指定这两种类型，返回 false。
// 工作原理：
// 1. 尝试获取 "Accept" 请求头。
// 2. 如果没有 Accept 头，默认客户端期望 JSON，返回 (ContentTypeJSON, true)。
// 3. 将 Accept 头的值按逗号分割成多个条目。
// 4. 遍历每个条目：
//    a. 处理空格和分号，提取媒体类型 (mediaType)，同 verifyJSONAcceptHeader。
//    b. 检查媒体类型是否是接受 JSON 的类型 ("*/*", "application/*", "application/json")。
//       如果是，立即返回 (ContentTypeJSON, true)。
//    c. 检查媒体类型是否是 "text/plain"。
//       如果是，立即返回 (ContentTypePlainText, true)。
// 5. 如果遍历完所有条目都没有找到明确接受 JSON 或 text/plain 的指令，说明无法确定客户端的偏好（或者 Accept 头无效），
//    返回 (ContentTypeJSON, false)，表示解析失败，但默认还是按 JSON 处理。
func parseJSONOrTextAcceptHeader(r *http.Request) (ContentType, bool) {
	// 尝试获取 "Accept" 请求头
	accept, ok := r.Header["Accept"]
	// 如果没有 Accept 头，默认返回 JSON，并标记为解析成功
	if !ok {
		return ContentTypeJSON, true
	}
	// 按逗号分割 Accept 头的值
	entries := strings.Split(accept[0], ",") // 只处理第一个 Accept 值
	// 遍历每个可接受的媒体类型条目
	for _, entry := range entries {
		// 去除首尾空格
		entry = strings.TrimSpace(entry)
		// 按分号分割，提取媒体类型部分
		parts := strings.Split(entry, ";")
		mediaType := strings.TrimSpace(parts[0])
		// 检查是否接受 JSON
		if mediaType == "*/*" || mediaType == "application/*" || mediaType == "application/json" {
			// 如果接受 JSON，返回 JSON 类型和解析成功标记
			return ContentTypeJSON, true
		}
		// 检查是否接受纯文本
		if mediaType == "text/plain" {
			// 如果接受纯文本，返回 PlainText 类型和解析成功标记
			return ContentTypePlainText, true
		}
	}
	// 如果遍历完都不能确定偏好，默认返回 JSON，但标记为解析失败
	return ContentTypeJSON, false
}

// ContentType 是一个自定义的整数类型，用于表示响应的内容格式。
// 使用自定义类型而不是直接用整数，可以提高代码的可读性和类型安全性。
type ContentType = int

// 定义内容类型的常量
const (
	// ContentTypeJSON 代表响应内容应该是 JSON 格式。
	ContentTypeJSON ContentType = iota // iota 是 Go 的常量计数器，这里会自动赋值为 0
	// ContentTypePlainText 代表响应内容应该是纯文本格式。
	ContentTypePlainText // iota 会自动递增，这里赋值为 1
)
