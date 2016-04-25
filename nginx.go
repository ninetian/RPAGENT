package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Nginx struct {
}

func (n *Nginx) Create(domainName string, IpAddr string, Port string, ssl bool, full bool, rootPath string) (m MessageModel) {

	defaultTmpl := "full.cdn.tmpl"

	if ssl {
		defaultTmpl = "ssl.cdn.tmpl"
	}

	if !full {
		defaultTmpl = "split.cdn.tmpl"
	}

	m = MessageModel{}
	m.Success = false

	template, err := n.getTemplateFile(defaultTmpl)

	if err != nil {
		m.Message = "Cannot read tmpl file: " + err.Error()
		log.Println(m.Message)

		return m
	}

	template = strings.Replace(template, "##DOMAIN##", domainName, -1)
	template = strings.Replace(template, "##IP##", IpAddr, -1)
	template = strings.Replace(template, "##PORT##", Port, -1)

	vhostConfName := fmt.Sprintf("%s.conf", domainName)
	vhostConfPath := filepath.Join(rootPath, vhostConfName)

	err = n.saveFile(vhostConfPath, template)

	if err != nil {
		m.Message = "Cannot save conf file to " + vhostConfPath + " error: " + err.Error()
		log.Println(m.Message)

		return m
	}

	m.Success = true
	m.Message = "Build Success: " + vhostConfPath
	log.Println(m.Message)

	err = n.reload()

	if err != nil {
		m.Message = m.Message + " | Nginx Reload: " + err.Error()
	}

	return m
}

func (n *Nginx) Delete(domainName string, rootPath string) (m MessageModel) {
	m = MessageModel{}
	m.Success = false

	vhostConfName := fmt.Sprintf("%s.conf", domainName)
	vhostConfPath := filepath.Join(rootPath, vhostConfName)

	err := n.deleteFile(vhostConfPath)

	if err != nil {
		m.Message = "File cannot be delete: " + vhostConfPath
		log.Println(m.Message)
		return
	}

	m.Success = true
	m.Message = "Domain deleted: " + domainName
	log.Println(m.Message)

	err = n.reload()

	if err != nil {
		m.Message = m.Message + " | Nginx Reload: " + err.Error()
	}

	return m
}

func (n *Nginx) List(rootPath string) (m MessageModelList) {
	m = MessageModelList{}
	m.Success = false
	m.Vhosts = []string{}

	filepath.Walk(rootPath, func(path string, f os.FileInfo, err error) error {
		m.Vhosts = append(m.Vhosts, path)
		return nil
	})

	m.Success = true
	m.Message = "List success"
	log.Println(m.Message)

	return m
}

func (n *Nginx) getTemplateFile(fileName string) (text string, err error) {

	if err != nil {
		return "", err
	}

	path := filepath.Join(config.Api.TemplatePath, fileName)
	b, err := ioutil.ReadFile(path)
	alltext := string(b)

	return alltext, err
}

func (n *Nginx) saveFile(filePath string, content string) (err error) {

	fileContent := []byte(content)
	err = ioutil.WriteFile(filePath, fileContent, 0644)

	return err
}

func (n *Nginx) deleteFile(fileName string) (err error) {

	isExists := n.fileExists(fileName)

	if !isExists {
		return err
	}

	err = os.Remove(fileName)

	return err
}

func (n *Nginx) fileExists(fileName string) bool {
	finfo, err := os.Stat(fileName)

	if err != nil {
		return false
	}

	return (finfo.IsDir() == false)
}

func (n *Nginx) reload() (err error) {

	log.Println("Reloading Nginx")
	cmd := exec.Command("service", "nginx", "reload")
	err = cmd.Run()

	return err
}
