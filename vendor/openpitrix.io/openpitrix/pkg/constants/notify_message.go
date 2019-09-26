// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package constants

import (
	"bytes"
	"fmt"
	"text/template"
)

const (
	en            = "en"
	zhCN          = "zh_cn"
	defaultLocale = zhCN
)

const EmailNotifyName = "email"

type EmailNotifyContent struct {
	Content string
}

type NotifyMessage struct {
	en   string
	zhCN string
}

type NotifyTitle struct {
	NotifyMessage
}

type NotifyContent struct {
	NotifyMessage
}

func (n *NotifyMessage) GetMessage(locale string, params ...interface{}) string {
	switch locale {
	case en:
		return fmt.Sprintf(n.en, params...)
	case zhCN:
		return fmt.Sprintf(n.zhCN, params...)
	default:
		return fmt.Sprintf(n.zhCN, params...)
	}
}

func (n *NotifyTitle) GetDefaultMessage(params ...interface{}) string {
	return n.GetMessage(defaultLocale, params...)
}

func (n *NotifyContent) GetDefaultMessage(params ...interface{}) string {
	t, _ := template.New(EmailNotifyName).Parse(EmailNotifyTemplate)

	b := bytes.NewBuffer([]byte{})
	emailContent := &EmailNotifyContent{
		Content: n.GetMessage(defaultLocale, params...),
	}

	t.Execute(b, emailContent)
	return b.String()
}

var (
	AdminInviteIsvNotifyTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】邀请您成为平台服务商",
		},
	}
	AdminInviteIsvNotifyContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: ` 
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">%s 您好</p>
                        <p class="line2">
                          <strong>%s</strong>邀请您入驻应用市场<strong>「%s」</strong>，成为优质服务商，为平台用户提供企业解决方案、产品和集成服务，共享快速收益。
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">接受邀请</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p>

                        <p class="line6">
                          用户名：<strong>%s</strong>
                        </p>

                        <p>
                          密码：<strong>%s</strong>
                        </p>

                         <p>
                          首次登陆后请修改密码。
                        </p>
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p> 
`,
		},
	}

	AdminInviteUserNotifyTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】邀请您成为平台用户",
		},
	}
	AdminInviteUserNotifyContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: `
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">Hi, %s</p>
                        <p class="line2">
                          <strong>%s</strong>邀请你加入<strong>「%s」</strong>，成为平台正式用户。
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">接受邀请</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p>

                        <p class="line6">
                          用户名：<strong>%s</strong>
                        </p>

                        <p>
                          密码：<strong>%s</strong>
                        </p>
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p>
`,
		},
	}

	IsvInviteMemberNotifyTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】邀请您加入 %s 平台",
		},
	}
	IsvInviteMemberNotifyContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: `  
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">%s 您好</p>
                        <p class="line2">
                          <strong>%s</strong>邀请您加入<strong>「%s」</strong>平台协同工作。
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">接受邀请</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p>

                        <p class="line6">
                          用户名：<strong>%s</strong>
                        </p>

                        <p>
                          密码：<strong>%s</strong>
                        </p>

                         <p>
                          首次登陆后请修改密码。
                        </p>
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p> 

`,
		},
	}

	SubmitVendorNotifyAdminTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】%s 应用服务商资质申请",
		},
	}
	SubmitVendorNotifyAdminContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: ` 
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">%s 您好</p>
                        <p class="line2">
                          收到 %s 应用服务商资质申请，请尽快完成审核。
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">审核</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p>

                        
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p> 

`,
		},
	}

	SubmitVendorNotifyIsvTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】已收到您的应用服务商资质申请",
		},
	}
	SubmitVendorNotifyIsvContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: ` 
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">%s 您好</p>
                        <p class="line2"> 
                        已收到您的应用服务商资质申请，我们会在3个工作日内完成审核，请您耐心等待。
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">查看申请</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p> 
                        
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p> 


`,
		},
	}

	PassVendorNotifyTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】您的 %s 应用服务商资质申请已通过",
		},
	}
	PassVendorNotifyContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: `  
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">%s 您好</p>
                        <p class="line2"> 
                        恭喜您，应用服务商资质申请通过审核，正式成为 %s 应用服务商。
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">查看详情</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p>

                        
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p> 
`,
		},
	}

	RejectVendorNotifyTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】已拒绝您的 %s 应用服务商资质申请",
		},
	}
	RejectVendorNotifyContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: `  
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">%s 您好</p>
                        <p class="line2"> 
                        您提交的 %s 应用服务商资质申请信息有误，请核对相关内容，完善申请后重新提交。
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">查看详情</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p>

                        
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p> 
`,
		},
	}

	SubmitAppVersionNotifyReviewerTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】%s 应用 %s 版本审核申请",
		},
	}
	SubmitAppVersionNotifyReviewerContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: ` 
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">%s 您好</p>
                        <p class="line2"> 
                        收到 %s 应用 %s 版本审核申请，请尽快完成审核。
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">查看详情</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p>

                        
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p>  
`,
		},
	}

	SubmitAppVersionNotifySubmitterTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】已收到您的 %s 应用 %s 版本审核申请",
		},
	}
	SubmitAppVersionNotifySubmitterContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: `
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">%s 您好</p>
                        <p class="line2"> 
                        已收到您的 %s 应用 %s 版本审核申请，请您耐心等待。
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">查看详情</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p>

                        
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p>  
`,
		},
	}

	PassAppVersionInfoNotifyTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】%s 应用 %s 版本通过应用信息审核",
		},
	}
	PassAppVersionInfoNotifyContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: ` 
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">%s 您好</p>
                        <p class="line2"> 
                        恭喜您，%s 应用 %s 版本已通过应用信息审核，等待平台商务审核。
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">查看详情</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p>

                        
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p>  

`,
		},
	}

	PassAppVersionBusinessNotifyTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】%s 应用 %s 版本通过平台商务审核",
		},
	}
	PassAppVersionBusinessNotifyContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: `
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">%s 您好</p>
                        <p class="line2"> 
                        恭喜您，%s 应用 %s 版本已通过平台商务审核，等待平台技术审核。
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">查看详情</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p>

                        
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p>  
`,
		},
	}

	PassAppVersionTechnicalNotifyTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】%s 应用 %s 版本通过平台技术审核",
		},
	}
	PassAppVersionTechnicalNotifyContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: `
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">%s 您好</p>
                        <p class="line2"> 
                       恭喜您，%s 应用 %s 版本已通过平台技术审核，请尽快完成应用版本上架。 
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">查看详情</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p>

                        
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p>  
`,
		},
	}

	RejectAppVersionInfoNotifyTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】%s 应用 %s 版本未通过应用信息审核",
		},
	}
	RejectAppVersionInfoNotifyContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: ` 
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">%s 您好</p>
                        <p class="line2"> 
                       您提交的 %s 应用 %s 版本未通过应用信息审核，请核对相关内容，完善后重新提交。
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">查看详情</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p>

                        
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p>   
`,
		},
	}

	RejectAppVersionBusinessNotifyTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】%s 应用 %s 版本未通过平台商务审核",
		},
	}
	RejectAppVersionBusinessNotifyContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: `
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">%s 您好</p>
                        <p class="line2"> 
                       您提交的 %s 应用 %s 版本未通过平台商务审核，请核对相关内容，完善后重新提交。
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">查看详情</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p>

                        
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p>   
`,
		},
	}

	RejectAppVersionTechnicalNotifyTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】%s 应用 %s 版本未通过平台技术审核",
		},
	}
	RejectAppVersionTechnicalNotifyContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: ` 
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">%s 您好</p>
                        <p class="line2"> 
                       您提交的 %s 应用 %s 版本未通过平台技术审核，请核对相关内容，完善后重新提交。
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">查看详情</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p>

                        
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p>

`,
		},
	}

	ReleaseAppVersionNotifyTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】%s 应用 %s 版本已上架",
		},
	}
	ReleaseAppVersionNotifyContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: `
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">%s 您好</p>
                        <p class="line2"> 
                       %s 应用 %s 版本已上架到应用市场。
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">查看详情</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p>

                        
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p>
`,
		},
	}

	SuspendAppVersionNotifyTitle = NotifyTitle{
		NotifyMessage: NotifyMessage{
			zhCN: "【%s】%s 应用 %s 版本已下架",
		},
	}
	SuspendAppVersionNotifyContent = NotifyContent{
		NotifyMessage: NotifyMessage{
			zhCN: `
                        <span class="platform">%s</span>
                        </p>
                        <p class="line1">%s 您好</p>
                        <p class="line2"> 
                       %s 应用 %s 版本已从应用市场下架。
                        </p>
                         
                        <p class="line3">
                          <a class="linkBtn" href="%s">查看详情</a>
                        </p>

                        <p class="line4">
                          如果按钮无法点击，请直接访问以下链接：
                        </p>

                        <p class="line5">
                          <a class="link" href="%s">%s</a>
                        </p>

                        
                        <hr />
                        <p class="gray">
                          * 此为系统邮件请勿回复
                        </p>
`,
		},
	}
)
