package models

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
)

// DecodeGroupIds 解析服务所属分组ID
func (this *Server) DecodeGroupIds() []int64 {
	if len(this.GroupIds) == 0 {
		return []int64{}
	}

	var result = []int64{}
	err := json.Unmarshal([]byte(this.GroupIds), &result)
	if err != nil {
		remotelogs.Error("Server.DecodeGroupIds", err.Error())
		// 忽略错误
	}
	return result
}

// DecodeHTTPPorts 获取HTTP所有端口
func (this *Server) DecodeHTTPPorts() (ports []int) {
	if len(this.Http) > 0 && this.Http != "null" {
		config := &serverconfigs.HTTPProtocolConfig{}
		err := json.Unmarshal([]byte(this.Http), config)
		if err != nil {
			return nil
		}
		err = config.Init()
		if err != nil {
			return nil
		}
		for _, listen := range config.Listen {
			for i := listen.MinPort; i <= listen.MaxPort; i++ {
				ports = append(ports, i)
			}
		}
	}
	return
}

// DecodeHTTPSPorts 获取HTTPS所有端口
func (this *Server) DecodeHTTPSPorts() (ports []int) {
	if len(this.Https) > 0 && this.Https != "null" {
		config := &serverconfigs.HTTPSProtocolConfig{}
		err := json.Unmarshal([]byte(this.Https), config)
		if err != nil {
			return nil
		}
		err = config.Init()
		if err != nil {
			return nil
		}
		for _, listen := range config.Listen {
			for i := listen.MinPort; i <= listen.MaxPort; i++ {
				ports = append(ports, i)
			}
		}
	}
	return
}

// DecodeTCPPorts 获取TCP所有端口
func (this *Server) DecodeTCPPorts() (ports []int) {
	if len(this.Tcp) > 0 && this.Tcp != "null" {
		config := &serverconfigs.TCPProtocolConfig{}
		err := json.Unmarshal([]byte(this.Tcp), config)
		if err != nil {
			return nil
		}
		err = config.Init()
		if err != nil {
			return nil
		}
		for _, listen := range config.Listen {
			for i := listen.MinPort; i <= listen.MaxPort; i++ {
				ports = append(ports, i)
			}
		}
	}
	return
}

// DecodeTLSPorts 获取TLS所有端口
func (this *Server) DecodeTLSPorts() (ports []int) {
	if len(this.Tls) > 0 && this.Tls != "null" {
		config := &serverconfigs.TLSProtocolConfig{}
		err := json.Unmarshal([]byte(this.Tls), config)
		if err != nil {
			return nil
		}
		err = config.Init()
		if err != nil {
			return nil
		}
		for _, listen := range config.Listen {
			for i := listen.MinPort; i <= listen.MaxPort; i++ {
				ports = append(ports, i)
			}
		}
	}
	return
}

// DecodeUDPPorts 获取UDP所有端口
func (this *Server) DecodeUDPPorts() (ports []int) {
	if len(this.Udp) > 0 && this.Udp != "null" {
		config := &serverconfigs.UDPProtocolConfig{}
		err := json.Unmarshal([]byte(this.Udp), config)
		if err != nil {
			return nil
		}
		err = config.Init()
		if err != nil {
			return nil
		}
		for _, listen := range config.Listen {
			for i := listen.MinPort; i <= listen.MaxPort; i++ {
				ports = append(ports, i)
			}
		}
	}
	return
}
