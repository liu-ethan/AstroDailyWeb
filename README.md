# Daily Fortune H5 (每日运势) 🔮

基于 Vue.js 和 Golang 开发的 H5 全栈应用。为用户提供专属的每日运势占卜、邮件订阅以及自动化定时推送服务。

## 🛠 技术栈

* **前端**: Vue.js (H5 端适配)
* **后端**: Golang
* **数据库**: MySQL 8.x
* **第三方服务**: 
  * 大模型 API (用于生成个性化运势)
  * 163 SMTP 服务 (用于发送验证码和每日运势邮件)

---

## ✨ 核心功能

* **🔐 账号认证**
  * 邮箱注册与验证码校验（SMTP 发送验证码）。
  * 邮箱 + 密码登录，JWT 有效期 30 分钟，Token 同步写入 Redis。
  * 支持邮箱验证码重置密码（密码长度不少于 8 位）。
* **🧾 用户资料**
  * 维护运势生成所需资料：出生日期、星座、性别、城市、职业。
  * 资料未完善时，接口会提示先补全资料。
* **🔯 运势生成**
  * 当日运势按“用户 + 日期”缓存至数据库，避免重复调用大模型。
  * 未命中缓存时调用大模型生成并落库。
* **✉️ 邮件订阅与定时任务**
  * 用户可订阅/取消订阅每日运势邮件。
  * **Cron Job**：每日 07:30 批量为订阅用户生成当日运势并发送邮件。
  * 自动清理 7 天前的历史运势数据。

---

## 🗄️ 数据库结构

项目依赖 3 张核心数据表：
1. `users`：存储用户账号、密码及订阅状态。
2. `fortunes`：存储用户每日的运势内容（联合唯一索引保证单人单日唯一）。
3. `verification_codes`：存储 SMTP 邮箱验证码及过期时间。

*(详细建表 SQL 请参考 `docs/sql/schema.sql`)*

---

## 🔌 API 接口概览

| 模块 | 接口路径 | 请求方式 | 鉴权 | 描述 |
| :--- | :--- | :---: | :---: | :--- |
| **Auth** | `/api/v1/auth/send-code` | POST | 否 | 发送邮箱验证码 (注册/重置密码) |
| **Auth** | `/api/v1/auth/register` | POST | 否 | 用户注册 |
| **Auth** | `/api/v1/auth/login` | POST | 否 | 用户登录 (返回 JWT) |
| **Auth** | `/api/v1/auth/reset-password` | POST | 否 | 重置密码 |
| **Fortune**| `/api/v1/fortune/today` | GET | 是 | 获取/生成今日运势 |
| **User** | `/api/v1/user/subscribe` | POST | 是 | 订阅每日运势邮件 |
| **User** | `/api/v1/user/unsubscribe`| POST | 是 | 取消订阅 |

---

## 🚀 快速开始

### 1. 环境准备
* Node.js (v16+)
* Go (v1.22+)
* MySQL (v8.0+)

### 2. 克隆项目
```bash
git clone [https://github.com/yourusername/daily-fortune-h5.git](https://github.com/yourusername/daily-fortune-h5.git)
cd daily-fortune-h5
```

### 3. 环境配置 (`.env`)
在后端根目录下创建 `.env` 文件，并配置以下环境变量：

```env
# 数据库配置
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=daily_fortune

# JWT 配置
JWT_SECRET=your_jwt_secret_key

# 163 SMTP 邮箱配置
SMTP_HOST=smtp.163.com
SMTP_PORT=465
SMTP_USER=your_email@163.com
SMTP_PASS=your_smtp_auth_code

# 大模型 API 配置
LLM_API_KEY=your_llm_api_key
LLM_API_URL=[https://api.example.com/v1/chat/completions](https://api.example.com/v1/chat/completions)
```

### 4. 启动服务

**后端 (Golang):**
```bash
cd backend
go mod tidy
go run main.go
# 服务默认运行在 http://localhost:9090
```

**前端 (Vue):**
```bash
cd frontend
npm install
npm run serve
# H5 页面默认运行在 http://localhost:9091
```

## 🐳 Docker 一键启动

使用 Docker Compose 构建并启动前后端、MySQL、Redis：

```bash
docker compose up -d --build
```

- 前端访问地址: http://localhost:9091
- 后端接口地址: http://localhost:9090
- 配置文件: `backend/config/config.docker.yaml`

如需启用邮件/大模型等能力，请在 `backend/config/config.docker.yaml` 中补充 SMTP 与 LLM 配置。

```bash
docker-compose build --build-arg GOPROXY=https://goproxy.cn,direct --no-cache backend
docker-compose up -d
docker exec -i astrodaily-mysql mysql -uroot astro_daily_web < create_db.sql
```
