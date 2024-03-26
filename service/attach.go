package service

import (
	"fmt"
	"ginchat/utils"
	"math/rand"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Upload(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.RespFail(c.Writer, err.Error())
		return
	}
	// 获取文件后缀
	suffix := ".png"
	tem := strings.Split(file.Filename, ".")
	if len(tem) > 1 {
		suffix = "." + tem[len(tem)-1]
	}
	// 生成文件名
	fileName := fmt.Sprintf("%d%04d%s", time.Now().Unix(), rand.Int31(), suffix)
	// 保存文件到指定目录
	err = c.SaveUploadedFile(file, "./asset/upload/"+fileName)
	if err != nil {
		utils.RespFail(c.Writer, err.Error())
		return
	}
	// 返回文件路径
	url := "./asset/upload/" + fileName
	utils.RespOk(c.Writer, url, "上传成功")
	// w := c.Writer
	// req := c.Request
	// srcFile, header, err := req.FormFile("file")
	// if err != nil {
	// 	utils.RespFail(w, err.Error())
	// }
	// suffix := ".png"
	// ofilName := header.Filename
	// tem := strings.Split(ofilName, ".")
	// if len(tem) > 1 {
	// 	suffix = "." + tem[len(tem)-1]
	// }
	// fileName := fmt.Sprintf("%d%04d%s", time.Now().Unix(), rand.Int31(), suffix)
	// dstFile, err := os.Create("./asset/upload/" + fileName)
	// if err != nil {
	// 	utils.RespFail(w, err.Error())
	// }
	// // 将源文件内容拷贝到目标文件
	// _, _ = io.Copy(dstFile, srcFile)
	// url := "./asset/upload/" + fileName
	// utils.RespOk(w, url, "上传成功")

}
