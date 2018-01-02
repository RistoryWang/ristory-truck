package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/ristorywang/ristory-truck/app/entity"
	"github.com/ristorywang/ristory-truck/app/libs"
	"github.com/ristorywang/ristory-truck/app/service"
	"strconv"
	"strings"
)

type ConfigController struct {
	BaseController
}

// 项目列表
func (this *ConfigController) List() {
	page, _ := strconv.Atoi(this.GetString("page"))
	if page < 1 {
		page = 1
	}

	count, _ := service.ConfigService.GetTotal()
	list, _ := service.ConfigService.GetList(page, this.pageSize)

	this.Data["count"] = count
	this.Data["list"] = list
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("ConfigController.List"), true).ToString()
	this.Data["pageTitle"] = "配置列表"
	this.display()
}

// 添加配置
func (this *ConfigController) Add() {

	if this.isPost() {
		p := &entity.Config{}
		p.Name = this.GetString("project_name")
		p.Type = this.GetString("project_type")
		p.AppProperties = this.GetString("app_properties")
		p.Log4jProperties = this.GetString("log4j_properties")
		p.ProdYml = this.GetString("prod_yml")
		p.Domain = this.GetString("project_domain")
		p.RepoUrl = this.GetString("repo_url")
		p.AgentId, _ = this.GetInt("agent_id")
		p.IgnoreList = this.GetString("ignore_list")
		p.BeforeShell = this.GetString("before_shell")
		p.AfterShell = this.GetString("after_shell")
		p.TaskReview, _ = this.GetInt("task_review")
		if v, _ := this.GetInt("create_verfile"); v > 0 {
			p.CreateVerfile = 1
		} else {
			p.CreateVerfile = 0
		}
		p.VerfilePath = strings.Replace(this.GetString("verfile_path"), ".", "", -1)

		if err := this.validConfig(p); err != nil {
			this.showMsg(err.Error(), MSG_ERR)
		}

		err := service.ConfigService.AddConfig(p)
		this.checkError(err)

		// 克隆仓库
		//		go service.ConfigService.CloneRepo(p.Id)

		service.ActionService.Add("add_config", this.auth.GetUserName(), "project", p.Id, "")

		this.redirect(beego.URLFor("ConfigController.List"))
	}

	agentList, err := service.ServerService.GetAgentList(1, -1)
	this.checkError(err)
	this.Data["pageTitle"] = "添加配置"
	this.Data["agentList"] = agentList
	this.display()
}

// 编辑配置
func (this *ConfigController) Edit() {
	id, _ := this.GetInt("id")
	p, err := service.ConfigService.GetConfig(id)
	this.checkError(err)

	if this.isPost() {
		p.Name = this.GetString("project_name")
		p.Type = this.GetString("project_type")
		p.AppProperties = this.GetString("app_properties")
		p.Log4jProperties = this.GetString("log4j_properties")
		p.ProdYml = this.GetString("prod_yml")
		p.AgentId, _ = this.GetInt("agent_id")
		p.IgnoreList = this.GetString("ignore_list")
		p.BeforeShell = this.GetString("before_shell")
		p.AfterShell = this.GetString("after_shell")
		p.TaskReview, _ = this.GetInt("task_review")
		if p.Status == -1 {
			p.RepoUrl = this.GetString("repo_url")
		}
		if v, _ := this.GetInt("create_verfile"); v > 0 {
			p.CreateVerfile = 1
		} else {
			p.CreateVerfile = 0
		}
		p.VerfilePath = strings.Replace(this.GetString("verfile_path"), ".", "", -1)

		if err := this.validConfig(p); err != nil {
			this.showMsg(err.Error(), MSG_ERR)
		}

		err := service.ConfigService.UpdateConfig(p, "Name", "AppProperties", "Log4jProperties", "ProdYml", "AgentId", "IgnoreList", "BeforeShell", "AfterShell", "RepoUrl", "CreateVerfile", "VerfilePath", "TaskReview")
		this.checkError(err)

		service.ActionService.Add("edit_config", this.auth.GetUserName(), "config", p.Id, "")

		this.redirect(beego.URLFor("ConfigController.List"))
	}

	agentList, err := service.ServerService.GetAgentList(1, -1)
	this.checkError(err)

	this.Data["project"] = p
	this.Data["agentList"] = agentList
	this.Data["pageTitle"] = "编辑配置"
	this.display()
}

// 删除项目
func (this *ConfigController) Del() {
	id, _ := this.GetInt("id")

	err := service.ConfigService.DeleteConfig(id)
	this.checkError(err)

	service.ActionService.Add("del_project", this.auth.GetUserName(), "project", id, "")

	this.redirect(beego.URLFor("ConfigController.List"))
}

// 重新克隆
func (this *ConfigController) Clone() {
	id, _ := this.GetInt("id")
	project, err := service.ConfigService.GetConfig(id)
	this.checkError(err)
	if project.Status != -1 {
		this.showMsg("只能对克隆失败的项目操作.", MSG_ERR)
	}

	project.Status = 0
	service.ConfigService.UpdateConfig(project, "Status")
	//	go service.ConfigService.CloneRepo(id)

	this.showMsg("", MSG_OK)
}

// 获取项目克隆状态
func (this *ConfigController) GetStatus() {
	id, _ := this.GetInt("id")
	project, _ := service.ConfigService.GetConfig(id)

	out := make(map[string]interface{})
	out["status"] = project.Status
	out["error"] = project.ErrorMsg

	this.jsonResult(out)
}

// 验证提交
func (this *ConfigController) validConfig(p *entity.Config) error {
	//else if p.RepoUrl == "" {
	//errorMsg = "请输入仓库地址"
	//}
	errorMsg := ""
	if p.Name == "" {
		errorMsg = "请输入项目名称"
	} else if p.Domain == "" {
		errorMsg = "请输入项目标识"
	} else if p.AgentId == 0 {
		errorMsg = "请选择跳板机"
	} else {
		agent, err := service.ServerService.GetServer(p.AgentId)
		if err != nil {
			return err
		}
		addr := fmt.Sprintf("%s:%d", agent.Ip, agent.SshPort)
		serv := libs.NewServerConn(addr, agent.SshUser, agent.SshKey)
		workPath := fmt.Sprintf("%s/%s", agent.WorkDir, p.Domain)

		if err := serv.TryConnect(); err != nil {
			errorMsg = "无法连接到跳板机: " + err.Error()
		} else if _, err := serv.RunCmd("mkdir -p " + workPath); err != nil {
			errorMsg = "无法创建跳板机工作目录: " + err.Error()
		}
		serv.Close()
	}

	if errorMsg != "" {
		return fmt.Errorf(errorMsg)
	}
	return nil
}
