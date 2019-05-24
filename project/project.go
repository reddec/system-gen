package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iancoleman/strcase"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"system-gen/templates"
)

const (
	File         = "system.json"
	GeneratedDir = "generated"
)

func Open(dirPath string) (*Project, error) {
	data, err := ioutil.ReadFile(dirPath + "/" + File)
	if err != nil {
		return nil, err
	}
	var p Project
	p.path = dirPath
	err = json.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}
	for _, s := range p.Services {
		s.Project = &p
	}
	for _, o := range p.OneShots {
		o.Project = &p
	}
	for _, t := range p.Timers {
		t.Project = &p
	}
	return &p, nil
}

type Project struct {
	Name     string     `json:"name"`
	Services []*Service `json:"services,omitempty"`
	Timers   []*Timer   `json:"timers,omitempty"`
	OneShots []*OneShot `json:"oneshots,omitempty"`
	path     string
}

func (p *Project) SaveAs(dirPath string) error {
	p.path = dirPath
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(p.File(), data, 0755)
}

func (p *Project) Save() error {
	return p.SaveAs(p.path)
}

func (p *Project) File() string {
	return p.path + "/" + File
}

func (p *Project) GeneratedDir() string {
	return p.path + "/" + GeneratedDir
}

func (p *Project) Slug() string {
	return strings.ToLower(strcase.ToKebab(p.Name))
}

func (p *Project) Service(srv *Service) {
	for i, s := range p.Services {
		if s.Name == srv.Name {
			p.Services[i] = srv
			return
		}
	}
	p.Services = append(p.Services, srv)
	srv.Project = p
}

func (p *Project) ServiceByName(name string) *Service {
	for _, s := range p.Services {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (p *Project) OneShot(srv *OneShot) {
	for i, s := range p.OneShots {
		if s.Name == srv.Name {
			p.OneShots[i] = srv
			return
		}
	}
	p.OneShots = append(p.OneShots, srv)
	srv.Project = p
}

func (p *Project) OneShotByName(name string) *OneShot {
	for _, s := range p.OneShots {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (p *Project) Timer(timer *Timer) {
	for i, s := range p.Timers {
		if s.Name == timer.Name {
			p.Timers[i] = timer
			return
		}
	}
	p.Timers = append(p.Timers, timer)
	timer.Project = p
}

func (p *Project) Render() (*Render, error) {
	data := &bytes.Buffer{}
	err := templates.ProjectUnitTemplate.Execute(data, p)
	return &Render{
		Content: data.Bytes(),
		dirPath: p.GeneratedDir(),
		name:    p.Slug() + ".service",
	}, err
}

func (p *Project) Renders() []Renderer {
	var ans []Renderer
	ans = append(ans, p, p.Installer(), p.UnInstaller())
	for _, srv := range p.Services {
		ans = append(ans, srv)
	}
	for _, srv := range p.OneShots {
		ans = append(ans, srv)
	}
	for _, timer := range p.Timers {
		ans = append(ans, timer)
	}
	return ans
}

func (p *Project) Installer() *Installer {
	return &Installer{project: p}
}
func (p *Project) UnInstaller() *UnInstaller {
	return &UnInstaller{project: p}
}

type Installer struct {
	project *Project
}

func (ins *Installer) Render() (*Render, error) {
	data := &bytes.Buffer{}
	err := templates.ProjectInstallerTemplate.Execute(data, ins.project)
	return &Render{
		Content: data.Bytes(),
		dirPath: ins.project.GeneratedDir(),
		name:    "install.sh",
		exec:    true,
	}, err
}

type UnInstaller struct {
	project *Project
}

func (ins *UnInstaller) Render() (*Render, error) {
	data := &bytes.Buffer{}
	err := templates.ProjectUnInstallerTemplate.Execute(data, ins.project)
	return &Render{
		Content: data.Bytes(),
		dirPath: ins.project.GeneratedDir(),
		name:    "uninstall.sh",
		exec:    true,
	}, err
}

type Render struct {
	Content []byte
	dirPath string
	name    string
	exec    bool
}

func (r *Render) Save() error {
	err := os.MkdirAll(r.dirPath, 0755)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(r.dirPath+"/"+r.name, r.Content, 0755)
	if err != nil {
		return err
	}
	if r.exec {
		return unix.Chmod(r.dirPath+"/"+r.name, 0755)
	}
	return nil
}

type Renderer interface {
	Render() (*Render, error)
}

type Service struct {
	Name        string            `json:"name"`
	ExecStart   string            `json:"exec_start"`
	Args        []string          `json:"args,omitempty"`
	Restart     string            `json:"restart"`
	RestartSec  int               `json:"restart_sec"`
	Environment map[string]string `json:"environment,omitempty"`
	Project     *Project          `json:"-"`
}

func (srv *Service) Render() (*Render, error) {
	data := &bytes.Buffer{}
	err := templates.ServiceUnitTemplate.Execute(data, srv)
	return &Render{
		Content: data.Bytes(),
		dirPath: srv.Project.GeneratedDir(),
		name:    srv.Slug() + ".service",
	}, err
}

func (srv *Service) Envs() []string {
	var ans []string
	for k, v := range srv.Environment {
		ans = append(ans, strconv.Quote(k+"="+v))
	}
	return ans
}

func (srv *Service) Binary() string {
	return buildExec(srv.ExecStart, srv.Args)
}

func (srv *Service) Slug() string {
	return srv.Project.Slug() + "-" + strings.ToLower(strcase.ToKebab(srv.Name))
}

type OneShot struct {
	Name        string            `json:"name"`
	ExecStart   string            `json:"exec_start"`
	Args        []string          `json:"args,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Project     *Project          `json:"-"`
}

func (srv *OneShot) Render() (*Render, error) {
	data := &bytes.Buffer{}
	err := templates.OneShotUnitTemplate.Execute(data, srv)
	return &Render{
		Content: data.Bytes(),
		dirPath: srv.Project.GeneratedDir(),
		name:    srv.Slug() + ".service",
	}, err
}

func (srv *OneShot) Envs() []string {
	var ans []string
	for k, v := range srv.Environment {
		ans = append(ans, strconv.Quote(k+"="+v))
	}
	return ans
}

func (srv *OneShot) Binary() string {
	return buildExec(srv.ExecStart, srv.Args)
}

func (srv *OneShot) Slug() string {
	return srv.Project.Slug() + "-" + strings.ToLower(strcase.ToKebab(srv.Name))
}

type Timer struct {
	Name     string   `json:"name"`
	Launch   string   `json:"launch"`
	Interval string   `json:"interval"`
	Project  *Project `json:"-"`
}

func (srv *Timer) Render() (*Render, error) {
	data := &bytes.Buffer{}
	err := templates.TimerUnitTemplate.Execute(data, srv)
	return &Render{
		Content: data.Bytes(),
		dirPath: srv.Project.GeneratedDir(),
		name:    srv.Slug() + ".timer",
	}, err
}

func (srv *Timer) Slug() string {
	return srv.Project.Slug() + "-" + strings.ToLower(strcase.ToKebab(srv.Name))
}

func (srv *Timer) Launcher() *OneShot {
	return srv.Project.OneShotByName(srv.Launch)
}

func buildExec(command string, args []string) string {
	var cmd = command
	if resolved, err := exec.LookPath(cmd); err == nil {
		fmt.Println("resolved binary as", resolved)
		if absPath, err := filepath.Abs(resolved); err == nil {
			fmt.Println("absolute path to binary is", absPath, "and will be used as executable")
			cmd = absPath
		} else {
			cmd = resolved
		}
	} else {
		fmt.Println("can't resolve binary:", err)
		fmt.Println("maybe path to binary is incorrect?")
	}
	if len(args) > 0 {
		var ans []string
		for _, arg := range args {
			ans = append(ans, strconv.Quote(arg))
		}
		cmd = cmd + " " + strings.Join(ans, " ")
	}
	return cmd
}
