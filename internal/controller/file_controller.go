package controller

import (
    "bytes"
    "errors"
    "image"
    "image/jpeg"
    _ "image/png"
    _ "image/gif"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "github.com/disintegration/imaging"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "go.mongodb.org/mongo-driver/bson"
    
    "roleplay/internal/repository"
)

const (
    maxAvatarSizeBytes = 5 * 1024 * 1024
)

// UploadAvatar 头像上传接口：校验、裁剪、压缩并保存原图与缩略图。
func UploadAvatar(c *gin.Context) {
    userId := c.GetString("userId")
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        respond(c, http.StatusBadRequest, "文件缺失", nil)
        return
    }
    if header.Size <= 0 || header.Size > maxAvatarSizeBytes {
        respond(c, http.StatusBadRequest, "文件大小不合法", nil)
        return
    }
    buf := new(bytes.Buffer)
    if _, err := buf.ReadFrom(file); err != nil {
        respond(c, http.StatusBadRequest, "读取文件失败", nil)
        return
    }
    data := buf.Bytes()
    if err := validateImageFile(data); err != nil {
        respond(c, http.StatusBadRequest, err.Error(), nil)
        return
    }
    // 解码
    img, _, err := image.Decode(bytes.NewReader(data))
    if err != nil {
        respond(c, http.StatusBadRequest, "图片解码失败", nil)
        return
    }
    // 裁剪为正方形，限制最大尺寸
    cropped := cropSquare(img)
    resized := imaging.Fit(cropped, 1024, 1024, imaging.Lanczos)
    thumb := imaging.Resize(cropped, 200, 200, imaging.Lanczos)

    // 编码JPEG
    var outFull, outThumb bytes.Buffer
    if err := jpeg.Encode(&outFull, resized, &jpeg.Options{Quality: 85}); err != nil {
        respond(c, http.StatusInternalServerError, "图片编码失败", nil)
        return
    }
    if err := jpeg.Encode(&outThumb, thumb, &jpeg.Options{Quality: 80}); err != nil {
        respond(c, http.StatusInternalServerError, "缩略图编码失败", nil)
        return
    }

    // 保存到磁盘
    if err := os.MkdirAll("uploads/avatars", 0755); err != nil {
        respond(c, http.StatusInternalServerError, "目录创建失败", nil)
        return
    }
    id := uuid.NewString()
    fullPath := filepath.Join("uploads", "avatars", id+".jpg")
    thumbPath := filepath.Join("uploads", "avatars", "thumb_"+id+".jpg")
    if err := os.WriteFile(fullPath, outFull.Bytes(), 0644); err != nil {
        respond(c, http.StatusInternalServerError, "保存失败", nil)
        return
    }
    if err := os.WriteFile(thumbPath, outThumb.Bytes(), 0644); err != nil {
        respond(c, http.StatusInternalServerError, "保存失败", nil)
        return
    }

    // 更新用户头像
    if err := updateUserAvatar(c, userId, "/static/"+fullPath, "/static/"+thumbPath); err != nil {
        // 不阻断返回
    }

    respond(c, http.StatusOK, "上传成功", gin.H{
        "avatar_url":   "/static/" + fullPath,
        "thumbnail_url": "/static/" + thumbPath,
        "uploaded_at":   time.Now().UTC(),
    })
}

func cropSquare(img image.Image) image.Image {
    b := img.Bounds()
    w, h := b.Dx(), b.Dy()
    size := w
    if h < w { size = h }
    return imaging.CropCenter(img, size, size)
}

func validateImageFile(data []byte) error {
    if len(data) < 8 {
        return errors.New("文件过小或损坏")
    }
    // 简单魔数校验（JPEG/PNG/GIF）
    if bytes.HasPrefix(data, []byte{0xFF, 0xD8}) { // JPEG
        return nil
    }
    if bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) { // PNG
        return nil
    }
    if bytes.HasPrefix(data, []byte{0x47, 0x49, 0x46}) { // GIF
        return nil
    }
    return errors.New("不支持的文件格式")
}

// updateUserAvatar 更新用户头像信息
func updateUserAvatar(c *gin.Context, userId, avatarURL, thumbnailURL string) error {
    update := bson.M{
        "$set": bson.M{
            "avatar":        avatarURL,
            "thumbnail":     thumbnailURL,
            "updatedAt":     time.Now(),
        },
    }
    
    _, err := repository.DB().Collection("users").UpdateOne(
        c, 
        bson.M{"userId": userId}, 
        update,
    )
    
    if err != nil {
        // 记录错误但不阻断流程
        return err
    }
    
    return nil
}

