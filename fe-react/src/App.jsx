import { useState, useEffect, useRef } from 'react'
import { MessageCircle, Plus, Trash2, Send, Bot, User, Sparkles, Menu } from 'lucide-react'

const API_BASE = '/api'

function App() {
  const [sessions, setSessions] = useState([])
  const [currentSession, setCurrentSession] = useState(null)
  const [messages, setMessages] = useState([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [connected, setConnected] = useState(false)
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const messagesEndRef = useRef(null)

  useEffect(() => {
    loadSessions()
    checkConnection()
  }, [])

  useEffect(() => {
    if (currentSession) {
      loadMessages(currentSession.id)
    }
  }, [currentSession])

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  const checkConnection = async () => {
    try {
      const res = await fetch(`${API_BASE}/ping`, { signal: AbortSignal.timeout(3000) })
      setConnected(res.ok)
    } catch {
      setConnected(false)
    }
  }

  const loadSessions = async () => {
    try {
      const res = await fetch(`${API_BASE}/sessions`)
      const data = await res.json()
      setSessions(data.sessions || [])
    } catch (error) {
      console.error('Failed to load sessions:', error)
    }
  }

  const loadMessages = async (sessionId) => {
    try {
      const res = await fetch(`${API_BASE}/messages?session_id=${sessionId}`)
      const data = await res.json()
      // 转换消息格式
      const formattedMessages = (data.messages || []).map(m => ({
        role: m.role,
        content: m.content,
        timestamp: new Date(m.created_at).getTime()
      }))
      setMessages(formattedMessages)
    } catch (error) {
      console.error('Failed to load messages:', error)
    }
  }

  const createSession = async () => {
    try {
      const res = await fetch(`${API_BASE}/sessions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title: '新会话' })
      })
      const session = await res.json()
      setSessions([session, ...sessions])
      setCurrentSession(session)
      setMessages([])
    } catch (error) {
      console.error('Failed to create session:', error)
    }
  }

  const selectSession = (session) => {
    setCurrentSession(session)
  }

  const deleteSession = async (e, sessionId) => {
    e.stopPropagation()
    try {
      await fetch(`${API_BASE}/sessions/${sessionId}`, { method: 'DELETE' })
      const newSessions = sessions.filter(s => s.id !== sessionId)
      setSessions(newSessions)
      if (currentSession?.id === sessionId) {
        setCurrentSession(newSessions[0] || null)
        setMessages([])
      }
    } catch (error) {
      console.error('Failed to delete session:', error)
    }
  }

  const clearSession = async () => {
    if (!currentSession) return
    // 删除并重新创建会话
    try {
      await fetch(`${API_BASE}/sessions/${currentSession.id}`, { method: 'DELETE' })
      const res = await fetch(`${API_BASE}/sessions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title: '新会话' })
      })
      const newSession = await res.json()
      const newSessions = sessions.map(s => 
        s.id === currentSession.id ? newSession : s
      )
      setSessions(newSessions)
      setCurrentSession(newSession)
      setMessages([])
    } catch (error) {
      console.error('Failed to clear session:', error)
    }
  }

  const sendMessage = async () => {
    if (!input.trim() || loading) return

    let session = currentSession
    if (!session) {
      // 自动创建会话
      try {
        const res = await fetch(`${API_BASE}/sessions`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ title: input.trim().slice(0, 20) })
        })
        session = await res.json()
        setSessions([session, ...sessions])
        setCurrentSession(session)
      } catch (error) {
        console.error('Failed to create session:', error)
        return
      }
    }

    const userMessage = {
      role: 'user',
      content: input.trim(),
      timestamp: Date.now()
    }

    setMessages(prev => [...prev, userMessage])
    setInput('')
    setLoading(true)

    try {
      const res = await fetch(`${API_BASE}/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session_id: session.id,
          message: userMessage.content,
          mode: 'text'
        })
      })

      const data = await res.json()
      const assistantMessage = {
        role: 'assistant',
        content: data.reply || '无响应',
        timestamp: Date.now()
      }

      setMessages(prev => [...prev, assistantMessage])
      
      // 刷新会话列表（更新标题和时间）
      loadSessions()

    } catch (error) {
      console.error('Error:', error)
      const errorMessage = {
        role: 'assistant',
        content: '抱歉，发生了错误，请稍后重试。',
        timestamp: Date.now()
      }
      setMessages(prev => [...prev, errorMessage])
    }

    setLoading(false)
  }

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      sendMessage()
    }
  }

  const suggestions = [
    { label: '介绍自己', prompt: '你好，请介绍一下你自己' },
    { label: '询问天气', prompt: '今天天气怎么样？' },
    { label: '讲个笑话', prompt: '给我讲个笑话' }
  ]

  return (
    <div className="h-screen flex bg-gray-50">
      {/* 侧边栏 */}
      <aside className={`${sidebarOpen ? 'w-72' : 'w-0'} bg-white border-r border-gray-200 flex flex-col transition-all duration-300 overflow-hidden`}>
        <div className="p-4 border-b border-gray-200 flex items-center justify-between">
          <h1 className="text-lg font-semibold text-gray-800 flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-primary-500" />
            AI 助手
          </h1>
          <button 
            onClick={createSession}
            className="p-2 rounded-lg bg-primary-500 text-white hover:bg-primary-600 transition-colors"
            title="新建会话"
          >
            <Plus className="w-5 h-5" />
          </button>
        </div>
        
        <div className="flex-1 overflow-y-auto p-2 space-y-1">
          {sessions.length === 0 ? (
            <div className="text-center py-8 text-gray-400">
              <MessageCircle className="w-12 h-12 mx-auto mb-3 opacity-50" />
              <p className="text-sm">暂无会话</p>
            </div>
          ) : (
            sessions.map(session => (
              <div
                key={session.id}
                onClick={() => selectSession(session)}
                className={`group p-3 rounded-xl cursor-pointer transition-all ${
                  currentSession?.id === session.id 
                    ? 'bg-primary-50 border border-primary-200' 
                    : 'hover:bg-gray-100'
                }`}
              >
                <div className="flex items-center gap-3">
                  <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${
                    currentSession?.id === session.id
                      ? 'bg-gradient-to-br from-primary-400 to-purple-500 text-white'
                      : 'bg-gradient-to-br from-gray-200 to-gray-300 text-gray-600'
                  }`}>
                    <Bot className="w-5 h-5" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-gray-800 truncate">{session.title}</p>
                    <p className="text-xs text-gray-400">
                      {new Date(session.updated_at).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })}
                    </p>
                  </div>
                  <button
                    onClick={(e) => deleteSession(e, session.id)}
                    className="opacity-0 group-hover:opacity-100 p-1.5 rounded-lg hover:bg-red-50 text-gray-400 hover:text-red-500 transition-all"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))
          )}
        </div>
        
        <div className="p-3 border-t border-gray-200">
          <div className="flex items-center gap-2 text-sm">
            <span className={`w-2 h-2 rounded-full ${connected ? 'bg-green-500' : 'bg-yellow-500'}`} />
            <span className="text-gray-500">{connected ? '已连接' : '离线'}</span>
          </div>
        </div>
      </aside>

      {/* 主聊天区域 */}
      <main className="flex-1 flex flex-col">
        {/* 顶部栏 */}
        <header className="px-6 py-4 bg-white border-b border-gray-200 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <button 
              onClick={() => setSidebarOpen(!sidebarOpen)}
              className="p-2 rounded-lg hover:bg-gray-100 text-gray-600 transition-colors lg:hidden"
            >
              <Menu className="w-5 h-5" />
            </button>
            <div>
              <h2 className="text-lg font-semibold text-gray-800">
                {currentSession?.title || '新会话'}
              </h2>
              {currentSession?.created_at && (
                <p className="text-xs text-gray-400">
                  {new Date(currentSession.created_at).toLocaleString('zh-CN', { 
                    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' 
                  })}
                </p>
              )}
            </div>
          </div>
          {messages.length > 0 && (
            <button 
              onClick={clearSession}
              className="px-3 py-1.5 text-sm text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
            >
              清空会话
            </button>
          )}
        </header>

        {/* 消息列表 */}
        <div className="flex-1 overflow-y-auto p-6 space-y-4">
          {messages.length === 0 ? (
            <div className="h-full flex flex-col items-center justify-center text-center px-4">
              <div className="w-24 h-24 rounded-3xl bg-gradient-to-br from-primary-400 to-purple-500 flex items-center justify-center text-white mb-6 shadow-lg">
                <Bot className="w-12 h-12" />
              </div>
              <h2 className="text-2xl font-bold text-gray-800 mb-2">你好，我是 AI 助手</h2>
              <p className="text-gray-500 mb-6">有什么我可以帮助你的吗？你可以尝试：</p>
              <div className="flex flex-wrap gap-3 justify-center">
                {suggestions.map((s, i) => (
                  <button
                    key={i}
                    onClick={() => setInput(s.prompt)}
                    className="px-5 py-2.5 bg-white border border-gray-200 text-gray-700 rounded-full text-sm hover:border-primary-500 hover:text-primary-600 hover:shadow-md transition-all"
                  >
                    {s.label}
                  </button>
                ))}
              </div>
            </div>
          ) : (
            messages.map((msg, i) => (
              <div 
                key={i}
                className={`flex gap-3 ${msg.role === 'user' ? 'flex-row-reverse' : ''} animate-fade-in`}
              >
                <div className={`w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0 ${
                  msg.role === 'user'
                    ? 'bg-gradient-to-br from-primary-400 to-purple-500 text-white'
                    : 'bg-gradient-to-br from-gray-100 to-gray-200 text-gray-600'
                }`}>
                  {msg.role === 'user' ? <User className="w-5 h-5" /> : <Bot className="w-5 h-5" />}
                </div>
                <div className={`max-w-[70%] ${msg.role === 'user' ? 'text-right' : ''}`}>
                  <div className={`px-4 py-3 rounded-2xl shadow-sm ${
                    msg.role === 'user'
                      ? 'bg-gradient-to-r from-primary-500 to-purple-500 text-white rounded-br-md'
                      : 'bg-white border border-gray-200 text-gray-800 rounded-bl-md'
                  }`}>
                    <p className="whitespace-pre-wrap leading-relaxed">{msg.content}</p>
                  </div>
                  <p className="text-xs text-gray-400 mt-1.5">
                    {new Date(msg.timestamp).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}
                  </p>
                </div>
              </div>
            ))
          )}
          
          {loading && (
            <div className="flex gap-3">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-gray-100 to-gray-200 text-gray-600 flex items-center justify-center">
                <Bot className="w-5 h-5" />
              </div>
              <div className="bg-white border border-gray-200 px-4 py-3 rounded-2xl rounded-bl-md shadow-sm">
                <div className="flex gap-1">
                  <span className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                  <span className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                  <span className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
                </div>
              </div>
            </div>
          )}
          
          <div ref={messagesEndRef} />
        </div>

        {/* 输入区域 */}
        <div className="px-6 py-4 bg-white border-t border-gray-200">
          <div className="flex items-end gap-3 bg-gray-50 border border-gray-200 rounded-2xl p-2 focus-within:border-primary-500 focus-within:ring-2 focus-within:ring-primary-100 transition-all">
            <textarea
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="输入消息..."
              rows={1}
              className="flex-1 bg-transparent border-none outline-none px-3 py-2 text-gray-800 resize-none max-h-32"
            />
            <button
              onClick={sendMessage}
              disabled={!input.trim() || loading}
              className="p-3 rounded-xl bg-gradient-to-r from-primary-500 to-purple-500 text-white hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
            >
              <Send className="w-5 h-5" />
            </button>
          </div>
          <div className="flex justify-between mt-2 text-xs text-gray-400">
            <span>Enter 发送，Shift+Enter 换行</span>
            <span>{input.length} 字符</span>
          </div>
        </div>
      </main>
    </div>
  )
}

export default App
