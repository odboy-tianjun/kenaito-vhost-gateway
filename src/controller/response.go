package controller

import (
	"encoding/json"
	"net/http"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`    // 状态码：200成功，500失败
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}

// Success 成功响应
func Success(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// Error 错误响应
func Error(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Code:    statusCode,
		Message: message,
		Data:    nil,
	})
}
