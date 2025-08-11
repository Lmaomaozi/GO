语言角色扮演后端（Go + MongoDB）

本项目实现用户系统、关系链（好友/群）、消息存储与查询，以及演绎房间消息接口。服务以 REST API 形式对外提供，存储使用 MongoDB。

运行步骤：
- 启动本地 MongoDB（默认 `mongodb://localhost:27017`，库名 `roleplay`）
- 修改 `configs/config.yaml`（JWT 秘钥、端口、数据库连接等）
- 执行：
  - `go mod tidy`
  - `go run ./cmd/server`

已实现的主要接口：
- 鉴权：POST /api/user/send_code，POST /api/user/login，POST /api/auth/refresh
- 用户：GET/PUT /api/user/me
- 关系链：好友申请/处理/列表、黑名单、群组增删成员/列表/详情
- 消息：POST /api/message/send，GET /api/message/history
- 房间：POST /api/room/join，GET /api/room/{id}/messages，POST /api/room/{id}/message

统一返回格式：`{ code, message, data }`

更多接口详情见 `docs/openapi.yaml`（中文）。

