package services

import (
	"encoding/json"
	"fmt"
	"os"
	"golang.org/x/crypto/bcrypt"
	"log"
)

// User 结构体表示用户信息
type User struct {
	Username string `json:"username"`
	Password string `json:"password"` // 存储的是bcrypt哈希值
}

// AuthService 结构体提供用户认证服务
type AuthService struct {
	usersFile string
	users     map[string]string // username -> hashed password
}

// NewAuthService 创建一个新的认证服务实例
func NewAuthService() *AuthService {
	service := &AuthService{
		usersFile: "users.json",
		users:     make(map[string]string),
	}
	
	// 加载现有用户
	service.loadUsers()
	
	// 如果没有用户，创建默认用户
	if len(service.users) == 0 {
		service.createDefaultUser()
	}
	
	return service
}

// loadUsers 从文件加载用户信息
func (s *AuthService) loadUsers() {
	// 检查用户文件是否存在
	if _, err := os.Stat(s.usersFile); os.IsNotExist(err) {
		log.Printf("用户文件不存在，将创建新文件: %s", s.usersFile)
		return
	}
	
	// 读取用户文件
	data, err := os.ReadFile(s.usersFile)
	if err != nil {
		log.Printf("读取用户文件失败: %v", err)
		return
	}
	
	// 解析用户数据
	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		log.Printf("解析用户数据失败: %v", err)
		return
	}
	
	// 填充用户映射
	for _, user := range users {
		s.users[user.Username] = user.Password
	}
	
	log.Printf("成功加载 %d 个用户", len(s.users))
}

// saveUsers 将用户信息保存到文件
func (s *AuthService) saveUsers() error {
	// 转换映射为切片
	var users []User
	for username, password := range s.users {
		users = append(users, User{
			Username: username,
			Password: password,
		})
	}
	
	// 序列化为JSON
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	
	// 写入文件
	return os.WriteFile(s.usersFile, data, 0600)
}

// createDefaultUser 创建默认管理员用户
func (s *AuthService) createDefaultUser() {
	username := "admin"
	password := "password"
	
	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("生成默认用户密码哈希失败: %v", err)
		return
	}
	
	// 添加到用户映射
	s.users[username] = string(hashedPassword)
	
	// 保存到文件
	if err := s.saveUsers(); err != nil {
		log.Printf("保存默认用户失败: %v", err)
		return
	}
	
	log.Printf("创建默认用户: %s", username)
}

// AuthenticateUser 验证用户凭据
func (s *AuthService) AuthenticateUser(username, password string) bool {
	// 检查用户是否存在
	hashedPassword, exists := s.users[username]
	if !exists {
		return false
	}
	
	// 验证密码
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// ChangePassword 修改用户密码
func (s *AuthService) ChangePassword(username, currentPassword, newPassword string) error {
	// 首先验证当前密码
	if !s.AuthenticateUser(username, currentPassword) {
		return fmt.Errorf("当前密码不正确")
	}
	
	// 生成新密码的哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("生成新密码哈希失败: %v", err)
	}
	
	// 更新密码
	s.users[username] = string(hashedPassword)
	
	// 保存到文件
	if err := s.saveUsers(); err != nil {
		return fmt.Errorf("保存用户信息失败: %v", err)
	}
	
	log.Printf("用户 %s 的密码已更新", username)
	return nil
}

// UserExists 检查用户是否存在
func (s *AuthService) UserExists(username string) bool {
	_, exists := s.users[username]
	return exists
}

// AddUser 添加新用户
func (s *AuthService) AddUser(username, password string) error {
	// 检查用户是否已存在
	if s.UserExists(username) {
		return fmt.Errorf("用户 %s 已存在", username)
	}
	
	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("生成密码哈希失败: %v", err)
	}
	
	// 添加用户
	s.users[username] = string(hashedPassword)
	
	// 保存到文件
	if err := s.saveUsers(); err != nil {
		return fmt.Errorf("保存用户信息失败: %v", err)
	}
	
	log.Printf("添加新用户: %s", username)
	return nil
}