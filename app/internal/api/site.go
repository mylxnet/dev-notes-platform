package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"dev-notes-platform/internal/validate"
)

// settingICPNumber 是站点配置表里备案号对应的键。
const settingICPNumber = "icp_number"

type settingsInput struct {
	ICPNumber string `json:"icp_number"`
}

// siteInfo 返回前台页脚需要的站点信息；未配置备案号时返回空串。
func (s *Server) siteInfo(c *gin.Context) {
	icp, err := s.store.GetSetting(c.Request.Context(), settingICPNumber)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "站点信息加载失败")
		return
	}
	respondOK(c, gin.H{"icp_number": icp})
}

// adminGetSettings 返回后台可编辑的站点配置。
func (s *Server) adminGetSettings(c *gin.Context) {
	icp, err := s.store.GetSetting(c.Request.Context(), settingICPNumber)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "站点设置加载失败")
		return
	}
	respondOK(c, gin.H{"icp_number": icp})
}

// adminUpdateSettings 保存站点配置；清空备案号后前台恢复占位文案。
func (s *Server) adminUpdateSettings(c *gin.Context) {
	var in settingsInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_BODY", "请求格式错误")
		return
	}
	in.ICPNumber = strings.TrimSpace(in.ICPNumber)
	if fieldErr := validate.ICPNumber(in.ICPNumber); fieldErr != nil {
		respondFieldError(c, fieldErr)
		return
	}
	if err := s.store.SetSetting(c.Request.Context(), settingICPNumber, in.ICPNumber); err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "站点设置保存失败")
		return
	}
	respondOK(c, gin.H{"icp_number": in.ICPNumber})
}
