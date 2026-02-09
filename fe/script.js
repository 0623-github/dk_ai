// 获取DOM元素
const userInput = document.getElementById('user-input');
const submitBtn = document.getElementById('submit-btn');
const responseBox = document.getElementById('response-box');

// API基础URL（根据后端服务配置）
const API_BASE_URL = 'http://localhost:9090';

// 初始化事件监听器
function initEventListeners() {
    // 提交按钮点击事件
    submitBtn.addEventListener('click', handleSubmit);
    
    // 回车键发送消息
    userInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSubmit();
        }
    });
}

// 处理提交事件
async function handleSubmit() {
    const message = userInput.value.trim();
    
    if (!message) return;
    
    // 清空输入框
    userInput.value = '';
    
    // 禁用按钮防止重复提交
    submitBtn.disabled = true;
    
    try {
        // 添加用户消息到界面
        addMessageToUI(message, 'user');
        
        // 调用API获取回答
        const response = await sendMessageToAPI(message);
        
        // 添加AI回答到界面
        addMessageToUI(response, 'system');
        
    } catch (error) {
        // 显示错误消息
        addMessageToUI('抱歉，处理您的请求时出现错误。请稍后重试。', 'system');
        console.error('API调用错误:', error);
    } finally {
        // 启用按钮
        submitBtn.disabled = false;
        // 滚动到底部
        scrollToBottom();
    }
}

// 发送消息到API
async function sendMessageToAPI(message) {
    // 添加loading状态
    addLoadingIndicator();
    
    try {
        // 构建请求数据
        const requestData = {
            message: message
        };
        
        // 发送POST请求到后端API
        const response = await fetch(`${API_BASE_URL}/chat`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestData)
        });
        
        // 移除loading状态
        removeLoadingIndicator();
        
        // 检查响应状态
        if (!response.ok) {
            throw new Error(`HTTP错误! 状态码: ${response.status}`);
        }
        
        // 解析JSON响应
        const data = await response.json();
        
        // 返回回答内容
        return data.response || data.text || '收到您的消息，但没有返回具体内容。';
        
    } catch (error) {
        // 移除loading状态
        removeLoadingIndicator();
        throw error;
    }
}

// 添加消息到UI界面
function addMessageToUI(message, type) {
    const messageDiv = document.createElement('div');
    messageDiv.className = `message ${type}`;
    
    const p = document.createElement('p');
    p.textContent = message;
    
    messageDiv.appendChild(p);
    responseBox.appendChild(messageDiv);
}

// 添加加载指示器
function addLoadingIndicator() {
    const loadingDiv = document.createElement('div');
    loadingDiv.id = 'loading-indicator';
    loadingDiv.className = 'message system';
    loadingDiv.innerHTML = '<p>正在思考中...</p>';
    
    responseBox.appendChild(loadingDiv);
    scrollToBottom();
}

// 移除加载指示器
function removeLoadingIndicator() {
    const loadingDiv = document.getElementById('loading-indicator');
    if (loadingDiv) {
        loadingDiv.remove();
    }
}

// 滚动到底部
function scrollToBottom() {
    const chatContainer = document.querySelector('.chat-container');
    chatContainer.scrollTop = chatContainer.scrollHeight;
}

// 检查后端API连接状态
async function checkAPIConnection() {
    try {
        // 检查后端ping端点
        const healthResponse = await fetch(`${API_BASE_URL}/ping`);
        return healthResponse.ok;
    } catch (error) {
        console.warn('无法连接到后端API，将使用模拟数据:', error);
        return false;
    }
}

// 模拟数据（当API不可用时使用）
function getMockResponse(message) {
    const mockResponses = {
        '你好': '你好！很高兴见到你，我是AI助手，有什么可以帮助你的？',
        '你叫什么名字': '我是AI聊天助手，可以回答你的问题。',
        '今天天气怎么样': '抱歉，我无法获取实时天气信息，但祝您有美好的一天！',
        '再见': '再见！有需要随时找我。'
    };
    
    return mockResponses[message] || '感谢您的提问！这是一个模拟回答。在实际环境中，您将收到来自AI模型的真实回复。';
}

// 初始化应用
async function initApp() {
    initEventListeners();
    
    // 检查API连接
    const isConnected = await checkAPIConnection();
    
    // 如果API不可用，使用模拟数据
    if (!isConnected) {
        // 替换sendMessageToAPI函数为模拟版本
        window.sendMessageToAPI = async (message) => {
            // 模拟网络延迟
            await new Promise(resolve => setTimeout(resolve, 1000));
            return getMockResponse(message);
        };
    }
    
    // 初始滚动到底部
    scrollToBottom();
}

// 页面加载完成后初始化
window.addEventListener('DOMContentLoaded', initApp);