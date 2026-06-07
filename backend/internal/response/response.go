package response

import "github.com/gin-gonic/gin"

func Success(c *gin.Context ,data any){

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": data,
	})
}


func Fail(c *gin.Context ,code int , msg string){
	c.JSON(code, gin.H{
		"code": code,
		"msg":  msg,
	})
}

func PageSuccess(c *gin.Context ,data any ,total int64 ,page int ,pageSize int){
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": data,
		"total": total,
		"page": page,
		"pageSize": pageSize,
	})
}