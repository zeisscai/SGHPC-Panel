package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	
	// 设置信任网络和转发头，确保在代理环境下能正确处理客户端IP等信息
	r.ForwardedByClientIP = true
	// 更安全的方式是明确指定可信代理IP范围，或在开发环境中设置为nil
	r.SetTrustedProxies(nil)
	
	// 设置静态文件服务
	r.Static("/static", "./web/static")
	
	// 添加对构建后前端资源的支持
	r.Static("/static/dist", "./web/static/dist")
	
	r.LoadHTMLGlob("web/templates/*")
	
	// 页面路由
	r.GET("/", indexHandler)
	
	// API路由
	api := r.Group("/api/v1")
	{
		api.GET("/nodes", getNodes)
		api.POST("/nodes", addNode)
		api.DELETE("/nodes/:id", removeNode)
		api.POST("/deploy", startDeployment)
		api.GET("/status", getDeploymentStatus)
		api.GET("/logs", getLogs)
	}
	
	// 修改监听地址为0.0.0.0，允许外部访问
	// 添加更详细的运行配置
	r.Run("0.0.0.0:8080")
}

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func getNodes(c *gin.Context) {
	// 获取节点列表
	c.JSON(http.StatusOK, gin.H{
		"nodes": []string{},
	})
}

func addNode(c *gin.Context) {
	// 添加节点
	c.JSON(http.StatusOK, gin.H{
		"message": "Node added successfully",
	})
}

func removeNode(c *gin.Context) {
	// 删除节点
	c.JSON(http.StatusOK, gin.H{
		"message": "Node removed successfully",
	})
}

func startDeployment(c *gin.Context) {
	// 开始部署
	c.JSON(http.StatusOK, gin.H{
		"message": "Deployment started",
	})
}

func getDeploymentStatus(c *gin.Context) {
	// 获取部署状态
	c.JSON(http.StatusOK, gin.H{
		"status": "running",
	})
}

func getLogs(c *gin.Context) {
	// 获取部署日志
	c.JSON(http.StatusOK, gin.H{
		"logs": []string{},
	})
}