package pref

import (
	"fmt"
	"github.com/gookit/color"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

var GlobalConfig ConcConfig
var GlobalDesc ConcDesc

type ConcConfig struct {
	AppVersion string `yaml:"conc_version"`
	AppLang string `yaml:"conc_lang"`

	OverwriteWhenSAM bool `yaml:"overwrite_when_sam"`
	MonitorPort string `yaml:"def_clnt_monitor_port"`
	ErrInfPort string `yaml:"def_clnt_errinfo_port"`

	ServerListenAddr string `yaml:"def_serv_listen_addr"`
	ServerListenPort string `yaml:"def_serv_listen_port"`
	ServerBufferSize int64 	`yaml:"def_serv_buffer_size"`
}

type ConcDesc struct {
	ErrNotSuffArgs string `yaml:"err_not_sufficient_arg"`

	InfTryConcHelp string `yaml:"inf_try_conc_help"`
	InfExit string `yaml:"inf_exit"`

	AppDesc string `yaml:"conc_description"`
	AppUsgTxtMacOS string `yaml:"conc_usage_txt_macos"`
	AppUsgTxtLinux string `yaml:"conc_usage_txt_linux"`
	AppUsgTxtWin string `yaml:"conc_usage_txt_win"`

	AppServerUsg string `yaml:"conc_server_usage"`
	AppServerAddrUsg string `yaml:"conc_server_a_usage"`
	AppServerPortUsg string `yaml:"conc_server_p_usage"`
	AppServerWkdrUsg string `yaml:"conc_server_w_usage"`
	AppServerBuffUsg string `yaml:"conc_server_b_usage"`

	AppClientUsg string `yaml:"conc_client_usage"`
	AppClientDesc string `yaml:"conc_client_desc"`
	AppClientMonUsg string `yaml:"conc_client_m_usage"`
	AppClientErrUsg string `yaml:"conc_client_e_usage"`
}

// 读取菜单文字文件
func (d *ConcDesc)ReadDesc () (*ConcDesc, bool) {
	var yamlFile []byte
	var err error
	var isDefault = true
	switch GlobalConfig.AppLang {
	case "zh-CN":
		yamlFile, err = ioutil.ReadFile("conf"+string(os.PathSeparator)+"conc_desc_zh-CN.yaml")
		if err != nil {
			break
		}
		err = yaml.Unmarshal(yamlFile, d)
		if err != nil {
			break
		}
		isDefault = false
	}

	// Default is English
	if isDefault {
		yamlFile, err = ioutil.ReadFile("conf"+string(os.PathSeparator)+"conc_desc_en-UK.yaml")
		if err != nil {
			HdrErr()
			fmt.Println("Unable to reach language document.")
			return nil, false
		}
		err = yaml.Unmarshal(yamlFile, d)
		if err != nil {
			HdrErr()
			fmt.Println("Bad language document format.")
			return nil, false
		}
	}
	return d, true
}

// 读取配置文件
func (c *ConcConfig)ReadConf() (*ConcConfig, bool) {
	yamlFile, err := ioutil.ReadFile("conf"+string(os.PathSeparator)+"conc_conf.yaml")
	if err != nil {
		HdrErr()
		fmt.Println("Unable to reach config document.")
		return nil, false
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		HdrErr()
		fmt.Println("Config document has bad format.")
		return nil, false
	}
	return c, true
}


func HdrErr() {
	color.Red.Print("CONC ERROR: ")
}
func HdrInf() {
	color.Green.Print("CONC INFO:  ")
}
func HdrWrn() {
	color.Yellow.Print("CONC WARNING:")
}

