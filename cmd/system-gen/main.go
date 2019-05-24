package main

import (
	"errors"
	"github.com/jessevdk/go-flags"
	"os"
	"system-gen/project"
)

type app struct {
	Init initProject `command:"init"`
	Add  struct {
		Service createService `command:"service"`
		Timer   createTimer   `command:"timer"`
		OneShot createOneShot `command:"oneshot"`
	} `command:"add"`
	Generate generate `command:"generate"`
}

func main() {
	var config app
	_, err := flags.Parse(&config)
	if err != nil {
		os.Exit(1)
	}
}

type initProject struct {
	Args struct {
		Project   string `required:"yes"`
		Directory string
	} `positional-args:"yes"`
}

func (ip *initProject) Execute([]string) error {
	p := &project.Project{
		Name: ip.Args.Project,
	}

	var directory = ip.Args.Directory
	if ip.Args.Directory == "" {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		dir = dir + "/" + p.Slug()
		directory = dir
	}
	return p.SaveAs(directory)
}

type createService struct {
	Once        bool              `long:"once" env:"ONCE" description:"Run service as 'oneshot'"`
	Restart     string            `long:"restart" env:"RESTART" description:"When to restart service" default:"always"`
	RestartSec  int               `long:"restart-sec" env:"RESTART_SEC" description:"Restart delay" default:"5"`
	Environment map[string]string `long:"environment" short:"e" env:"ENVIRONMENT" description:"Additional environment variables"`
	Args        struct {
		ServiceName string `required:"yes"`
		Command     string `required:"yes"`
		Args        []string
	} `positional-args:"yes"`
}

func (cb *createService) Execute(args []string) error {
	p, err := project.Open(".")
	if err != nil {
		return err
	}
	srv := &project.Service{
		Name:        cb.Args.ServiceName,
		ExecStart:   cb.Args.Command,
		Args:        cb.Args.Args,
		Restart:     cb.Restart,
		RestartSec:  cb.RestartSec,
		Environment: cb.Environment,
	}
	p.Service(srv)
	return p.Save()

}

type createOneShot struct {
	Environment map[string]string `long:"environment" short:"e" env:"ENVIRONMENT" description:"Additional environment variables"`
	Args        struct {
		Name    string `required:"yes"`
		Command string `required:"yes"`
		Args    []string
	} `positional-args:"yes"`
}

func (cb *createOneShot) Execute(args []string) error {
	p, err := project.Open(".")
	if err != nil {
		return err
	}

	srv := &project.OneShot{
		Name:        cb.Args.Name,
		ExecStart:   cb.Args.Command,
		Args:        cb.Args.Args,
		Environment: cb.Environment,
	}
	p.OneShot(srv)
	return p.Save()

}

type createTimer struct {
	Name string `long:"name" env:"NAME" description:"Custom name for timer. By default - launcher name with -timer-interval suffix"`
	Args struct {
		Launch   string `required:"yes" description:"service name that should be launched"`
		Interval string `required:"yes" description:"interval to start service between inactivity (suffixes: us, ms, s, m, h, d, w, M, y). Can be combined"`
	} `positional-args:"yes"`
}

func (ct *createTimer) Execute(args []string) error {
	p, err := project.Open(".")
	if err != nil {
		return err
	}
	srv := p.OneShotByName(ct.Args.Launch)
	if srv == nil {
		return errors.New("one-shot service " + ct.Args.Launch + " that should be launched not found")
	}
	name := ct.Name
	if name == "" {
		name = ct.Args.Launch + "-timer-" + ct.Args.Interval
	}
	p.Timer(&project.Timer{
		Interval: ct.Args.Interval,
		Launch:   ct.Args.Launch,
		Name:     name,
	})
	return p.Save()
}

type generate struct {
}

func (*generate) Execute(args []string) error {
	p, err := project.Open(".")
	if err != nil {
		return err
	}

	err = os.RemoveAll(p.GeneratedDir())
	if err != nil {
		return err
	}

	for _, srv := range p.Renders() {
		render, err := srv.Render()
		if err != nil {
			return err
		}
		err = render.Save()
		if err != nil {
			return err
		}
	}
	return nil
}
