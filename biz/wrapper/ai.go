package wrapper

import (
	"context"
	"time"

	"github.com/0623-github/dk_ai/lib/ai"
	"github.com/0623-github/dk_ai/lib/ai/ollama"
	"github.com/0623-github/dk_ai/lib/db"
)

type ChatRequest struct {
	SessionID string `json:"session_id"`
	User      string `json:"user"`
	Message   string `json:"message"`
	Mode      string `json:"mode"`
}

type ChatResp struct {
	SessionID string `json:"session_id"`
	Reply     string `json:"reply"`
	Timestamp int64  `json:"timestamp"`
}

type SessionResp struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MessageResp struct {
	ID        int64     `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Wrapper interface {
	// 会话管理
	CreateSession(title string) (*SessionResp, error)
	GetSession(id string) (*SessionResp, error)
	ListSessions(limit int) ([]*SessionResp, error)
	UpdateSessionTitle(id, title string) error
	DeleteSession(id string) error

	// 消息管理
	GetMessages(sessionID string, limit int) ([]*MessageResp, error)

	// 聊天
	Chat(ctx context.Context, req ChatRequest) (ChatResp, error)
	ChatStream(ctx context.Context, req ChatRequest, callback func(string, bool)) error
}

type Impl struct {
	AIClient ai.AI
	DB       *db.DB
}

func NewImpl(ctx context.Context, database *db.DB) *Impl {
	aiClient := ollama.NewImpl(ctx)
	return &Impl{
		AIClient: aiClient,
		DB:       database,
	}
}

// CreateSession 创建新会话
func (w *Impl) CreateSession(title string) (*SessionResp, error) {
	id := generateSessionID()
	if title == "" {
		title = "新会话"
	}
	session, err := w.DB.CreateSession(id, title)
	if err != nil {
		return nil, err
	}
	return &SessionResp{
		ID:        session.ID,
		Title:     session.Title,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}, nil
}

// GetSession 获取会话
func (w *Impl) GetSession(id string) (*SessionResp, error) {
	session, err := w.DB.GetSession(id)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}
	return &SessionResp{
		ID:        session.ID,
		Title:     session.Title,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}, nil
}

// ListSessions 列出所有会话
func (w *Impl) ListSessions(limit int) ([]*SessionResp, error) {
	sessions, err := w.DB.ListSessions(limit)
	if err != nil {
		return nil, err
	}
	var result []*SessionResp
	for _, s := range sessions {
		result = append(result, &SessionResp{
			ID:        s.ID,
			Title:     s.Title,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		})
	}
	return result, nil
}

// UpdateSessionTitle 更新会话标题
func (w *Impl) UpdateSessionTitle(id, title string) error {
	return w.DB.UpdateSessionTitle(id, title)
}

// DeleteSession 删除会话
func (w *Impl) DeleteSession(id string) error {
	return w.DB.DeleteSession(id)
}

// GetMessages 获取会话消息
func (w *Impl) GetMessages(sessionID string, limit int) ([]*MessageResp, error) {
	messages, err := w.DB.GetMessagesBySession(sessionID, limit)
	if err != nil {
		return nil, err
	}
	var result []*MessageResp
	for _, m := range messages {
		result = append(result, &MessageResp{
			ID:        m.ID,
			SessionID: m.SessionID,
			Role:      m.Role,
			Content:   m.Content,
			CreatedAt: m.CreatedAt,
		})
	}
	return result, nil
}

// Chat 普通聊天
func (w *Impl) Chat(ctx context.Context, req ChatRequest) (resp ChatResp, err error) {
	// 确保会话存在
	session, err := w.DB.GetSession(req.SessionID)
	if err != nil {
		return resp, err
	}
	if session == nil {
		// 自动创建会话
		_, err = w.DB.CreateSession(req.SessionID, "新会话")
		if err != nil {
			return resp, err
		}
	}

	// 保存用户消息
	_, err = w.DB.CreateMessage(req.SessionID, "user", req.Message)
	if err != nil {
		return resp, err
	}

	// 获取历史消息作为上下文
	history, err := w.DB.GetMessagesBySession(req.SessionID, 20)
	if err != nil {
		return resp, err
	}

	// 构建 AI 请求（包含历史上下文）
	aiReq := ai.ChatRequest{
		User:    req.User,
		Message: req.Message,
	}
	// 转换历史消息为 AI 上下文
	if len(history) > 0 {
		aiReq.History = make([]ai.ChatMessage, 0, len(history))
		for _, msg := range history {
			aiReq.History = append(aiReq.History, ai.ChatMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// 调用 AI
	reply, err := w.AIClient.Chat(ctx, aiReq)
	if err != nil {
		return resp, err
	}

	// 保存 AI 回复
	_, err = w.DB.CreateMessage(req.SessionID, "assistant", reply)
	if err != nil {
		return resp, err
	}

	// 更新会话时间
	w.DB.UpdateSessionTimestamp(req.SessionID)

	resp = ChatResp{
		SessionID: req.SessionID,
		Reply:     reply,
		Timestamp: time.Now().Unix(),
	}
	return resp, nil
}

// ChatStream 流式聊天
func (w *Impl) ChatStream(ctx context.Context, req ChatRequest, callback func(string, bool)) error {
	// 确保会话存在
	session, err := w.DB.GetSession(req.SessionID)
	if err != nil {
		return err
	}
	if session == nil {
		_, err = w.DB.CreateSession(req.SessionID, "新会话")
		if err != nil {
			return err
		}
	}

	// 保存用户消息
	_, err = w.DB.CreateMessage(req.SessionID, "user", req.Message)
	if err != nil {
		return err
	}

	// 获取历史消息
	history, err := w.DB.GetMessagesBySession(req.SessionID, 20)
	if err != nil {
		return err
	}

	// 构建 AI 请求
	aiReq := ai.ChatRequest{
		User:    req.User,
		Message: req.Message,
	}
	if len(history) > 0 {
		aiReq.History = make([]ai.ChatMessage, 0, len(history))
		for _, msg := range history {
			aiReq.History = append(aiReq.History, ai.ChatMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// 收集完整回复
	var fullReply string
	err = w.AIClient.ChatStream(ctx, aiReq, func(chunk string, done bool) {
		fullReply += chunk
		callback(chunk, done)
	})
	if err != nil {
		return err
	}

	// 保存 AI 回复
	_, err = w.DB.CreateMessage(req.SessionID, "assistant", fullReply)
	if err != nil {
		return err
	}

	// 更新会话时间
	w.DB.UpdateSessionTimestamp(req.SessionID)

	return nil
}

// generateSessionID 生成会话 ID
func generateSessionID() string {
	return "session_" + time.Now().Format("20060102150405") + "_" + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
