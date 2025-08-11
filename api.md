数据库结构设计
我们还需要一个会员管理系统，至少是个能看的系统再说
核心实体表
1. 用户表 (users)
{
  "_id": "ObjectId",
  "phone": "string", // 手机号
  "nickname": "string", // 昵称
  "avatar": "string", // 头像URL
  "userId": "string", // 内部的用户 ID，不显示给用户
  "userOpenId": "string", // 显示给用户的 ID，除了显示和找人没别的用处
  "wordCount": "int", // 总字数
  "coinCount": "int", // A币数
  "gemCount": "int", // A钻数
  "createdAt": "datetime",
  "updatedAt": "datetime",
  "deletedAt": "datetime"
}
2. 用户皮 (costumes)
用户创建的一个扮演角色
{
  "_id": "ObjectId",
  "nickname": "string", // 昵称
  "avatar": "string", // 头像URL
  "userId": "string", // 内部的用户 ID，不显示给用户
  "characterId": "string", // 源角色 ID
  "userOpenId": "string", // 显示给用户的 ID，除了显示和找人没别的用处
  "createdAt": "datetime",
  "updatedAt": "datetime",
  "deletedAt": "datetime"
}
3. 剧本表 (backstories)
{
  "_id": "ObjectId",
  "title": "string",           // 标题
  "subtitle": "string",        // 副标题
  "cover": [{"url": "string", "type": "string" }],    // 封面
  "content": "string",         // 故事内容
  "authorName": "string", // 作者
  "authorUserId": "string", // 作者
  "tags": ["string"],          // 标签列表
  "characters": [              // 角色列表
    {
      "character_id": "string",
      "name": "string",
      "avatar": "string",
      "illustration": "string", // 立绘
      "story": "string"         // 角色故事
    }
  ],
  "viewCount": "int",         // 点击量
  "likeCount": "int",         // 点赞量
  "passReview": "string", // 是否审核通过
  "reviewedAt": "datetime", // 审核通过时间
  "reviewedByUserId": "string", // 审核员
  "createdAt": "datetime",
  "updatedAt": "datetime",
  "deletedAt": "datetime"
}
3. 招募表 (recruits)
@郭睿明 这里还需要重新考虑下，对一下群戏的逻辑。
{
  "_id": "ObjectId",
  "title": "string",                    // 招募标题
  "backstoryId": "ObjectId",           // 关联剧本
  "creatorId": "ObjectId",             // 发布者
  "mode": "string", // 对戏、群戏
  "myCharacters": ["string"],          // 我的角色ID列表
  "targetCharacters": ["string"],      // 对方角色ID列表
  "customContent": "string",           // 自定义剧情(剧情模式)
  "customCharacters": [                // 自定义角色(剧情模式)
    {
      "characterId": "string",
      "name": "string",
      "avatar": "string",
      "illustration": "string",
      "background": "string"
    }
  ],
  "status": "string",                   // active/completed/cancelled
  "createdAt": "datetime",
  "updatedAt": "datetime",
  "deletedAt": "datetime"
}
4. 演绎剧场表（群聊房间） (theaters)
{
  "_id": "ObjectId",
  "recruitId": "ObjectId",             // 关联招募
  "backstoryId": "ObjectId",           // 关联剧本
  "title": "string",
  "subtitle": "string",
  "mode": "string",                     // couple/crowd/drama
  "backgroundStory": "string",         // 背景故事
  "participants": [                     // 参与者
    {
      "userId": "ObjectId",
      "costumeId": "string",
      "costumeName": "string",
      "avatar": "string",
      "joinTime": "datetime",
      "messageCount": "int"            // 发言数量
    }
  ],
  "status": "string",                   // active/completed/archived
  "createdAt": "datetime",
  "updatedAt": "datetime",
  "deletedAt": "datetime"
}
5. 消息表 (messages)
{
  "_id": "ObjectId",
  "roomId": "ObjectId",                // 房间ID
  "senderUserId": "ObjectId",              // 发送者用户ID
  "messageType": "string",             // character/user/system/start_block
  "element": {
    "type": "string",
    "data": {} // 任意对象，具体类型由 type 确定
  },
  "characterInfo": { // 角色信息(皮上消息)
    "characterId": "string",
    "name": "string", 
    "avatar": "string"
  },
  "createdAt": "datetime",
  "updatedAt": "datetime",
  "deletedAt": "datetime"
}
6. 戏文表 (cassettes，录影带，VCR 的 C 就是这个)
{
  "_id": "ObjectId",
  "title": "string",                    // 戏文标题
  "description": "string",              // 戏文描述
  "backstoryId": "ObjectId",           // 关联剧本
  "roomId": "ObjectId",                // 关联房间
  "creatorId": "ObjectId",             // 发布者
  "participants": [                     // 参与者信息
    {
      "userId": "ObjectId",
      "characterId": "string",
      "characterName": "string"
    }
  ],
  "messageIds": ["ObjectId"],          // 选中的消息ID列表
  "likeCount": "int",                  // 点赞数
  "viewCount": "int",                  // 浏览数
  "createdAt": "datetime",
  "updatedAt": "datetime", 
  "deletedAt": "datetime"
}
7. 其他辅助表
@郭睿明 取消点赞怎么办？
还有好友表和黑名单表
// 点赞表 (likes)
{
  "_id": "ObjectId",
  "userId": "ObjectId",
  "targetType": "string",              // backstory/record
  "targetId": "ObjectId",
  "createdAt": "datetime"
}

// 水区帖子表 (posts)
{
  "_id": "ObjectId", 
  "title": "string",
  "content": "string",
  "authorId": "ObjectId",
  "replyCount": "int",
  "likeCount": "int",
  "createdAt": "datetime",
  "updatedAt": "datetime",
  "deletedAt": "datetime"
}

// 反馈表 (feedbacks)
{
  "_id": "ObjectId",
  "userId": "ObjectId", 
  "content": "string",
  "type": "string",                     // feedback/report
  "status": "string",                   // pending/processing/resolved
  "createdAt": "datetime",
  "updatedAt": "datetime"
}
活动用
每日签到活动

API设计
通用返回格式
{
  "code": 200,
  "message": "success", 
  "data": {}
}
错误码规范
- 200: 成功
- 400: 参数错误
- 401: 未授权
- 403: 禁止访问
- 404: 资源不存在
- 500: 服务器错误
分页系统
对于需要分页加载的内容，必然有一个连续且单调的属性，比如 msgSeq 序列号；分页的请求格式为：
{
    "lastSeq": "string",
    "limit": "number"
}
API接口设计
1. 剧本系统 (backstory)
GET /api/backstory/list
- 功能：查询剧本列表
- 鉴权：任何人
- 参数：
{
  "page": 1,
  "size": 20,
  "tags": ["string"],              // 可选，标签筛选
  "keyword": "string"              // 可选，关键词搜索
}
- 返回：
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 100,
    "list": [
      {
        "id": "string",
        "title": "string",
        "subtitle": "string", 
        "cover_square": "string",
        "author": "string",
        "tags": ["string"],
        "view_count": 1000,
        "like_count": 50,
        "created_at": "2024-01-01T10:00:00Z"
      }
    ]
  }
}
GET /api/backstory/detail/{id}
- 功能：查询剧本详情
- 鉴权：任何人
- 返回：
{
  "code": 200,
  "message": "success", 
  "data": {
    "id": "string",
    "title": "string",
    "subtitle": "string",
    "cover_rect": "string",
    "cover_square": "string", 
    "content": "string",
    "author": "string",
    "tags": ["string"],
    "characters": [
      {
        "character_id": "string",
        "name": "string",
        "avatar": "string",
        "illustration": "string",
        "background": "string"
      }
    ],
    "view_count": 1000,
    "like_count": 50,
    "created_at": "2024-01-01T10:00:00Z"
  }
}
投递剧本

审核剧本通过 / 不通过


2. 戏文系统 (record)
GET /api/record/list
- 功能：查询戏文列表
- 鉴权：任何人
- 参数：同剧本列表
- 返回：
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 100,
    "list": [
      {
        "id": "string",
        "title": "string", 
        "description": "string",
        "backstory_title": "string",
        "participants": [
          {
            "user_nickname": "string",
            "character_name": "string"
          }
        ],
        "like_count": 20,
        "view_count": 200,
        "created_at": "2024-01-01T10:00:00Z"
      }
    ]
  }
}
GET /api/record/detail/{id}
- 功能：查询戏文详情(不含消息列表)
- 鉴权：任何人
- 返回：戏文基本信息
GET /api/record/message/{id}
- 功能：获取戏文的消息列表
- 鉴权：任何人
- 返回：
{
  "code": 200,
  "message": "success",
  "data": {
    "messages": [
      {
        "id": "string",
        "message_type": "character",
        "content": "string",
        "character_info": {
          "character_name": "string",
          "character_avatar": "string"
        },
        "created_at": "2024-01-01T10:00:00Z"
      }
    ]
  }
}
3. 招募系统 (recruit)
GET /api/recruit/list
- 功能：查询招募列表
- 鉴权：任何人
- 参数：分页参数
- 返回：招募列表
POST /api/recruit/create
- 功能：发布招募
- 鉴权：登录用户
- 参数：
{
  "backstory_id": "string",
  "mode": "couple",                    // couple/crowd/drama
  "myCharacters": ["string"],
  "targetCharacters": ["string"],
  "title": "string",                   // 可选，自定义标题
  "customContent": "string",          // 剧情模式的自定义内容
  "customCharacters": []              // 剧情模式的自定义角色
}
返回值：
{
    "id": "string"
}
DELETE /api/recruit/{id}
- 功能：删除招募
- 鉴权：发布者本人或管理员
- 返回：操作结果
4. 演绎系统 (room)
POST /api/room/join
- 功能：加入演绎
- 鉴权：登录用户
- 参数：
{
  "recruit_id": "string",
  "character_id": "string"             // 选择的角色
}
返回：操作结果
GET /api/room/{id}/messages
- 功能：获取房间消息
- 鉴权：房间参与者
- 参数：分页参数
- 返回：消息列表
POST /api/room/{id}/message
- 功能：发送消息
- 鉴权：房间参与者
- 参数：
{
  "content": "string",
  "messageType": "character",         // character/user
  "character_id": "string"             // 皮上消息时需要
}
5. 用户系统 (user)
POST /api/user/login
- 功能：用户登录
- 参数：
{
  "phone": "string",
  "code": "string"                     // 验证码
}
POST /api/user/send_code
- 功能：发送验证码
- 参数：
{
  "phone": "string"
}
6. 其他接口
POST /api/like
- 功能：点赞/取消点赞
- 鉴权：登录用户
- 参数：
{
  "target_type": "backstory",          // backstory/record
  "target_id": "string"
}
POST /api/feedback
- 功能：提交反馈
- 鉴权：登录用户
- 参数：
{
  "content": "string",
  "type": "feedback" // feedback/report
}