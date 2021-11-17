package main

import (
	"bytes"
	"errors"
	"github.com/ZJGSU-ACM/GoOnlineJudge/model"
	"github.com/ZJGSU-ACM/RunServer/config"
	"github.com/ZJGSU-ACM/vjudger"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var ErrCompile = errors.New("compile error")

const ZJGSUToken = "ZJGSU"

type ZJGSUJudger struct {
	token   string
	workdir string
	time    int
	mem     int
} //ZJGSUJudger implements vjudger.Vjudger interface.

func (z *ZJGSUJudger) Init(user vjudger.UserInterface) error {
	z.token = ZJGSUToken

	z.workdir = runPath + "/" + strconv.Itoa(user.GetSid()) + "/" + strconv.Itoa(user.GetVid())
	logger.Println("workdir is ", z.workdir)

	log.Println(z.workdir)

	cmd := exec.Command("mkdir", "-p", z.workdir)
	err := cmd.Run()
	if err != nil {
		log.Printf("mkdir workdir finished with error: %v", err)
	}

	z.files(user, z.workdir)
	return nil
}

func (z *ZJGSUJudger) Match(token string) bool {

	if ZJGSUToken == token || token == "" {
		return true
	}
	return false
}

//Get problem Info
func (z *ZJGSUJudger) Login(user vjudger.UserInterface) error {
	proModel := &model.ProblemModel{}
	logger.Println(user.GetVid())
	pro, _ := proModel.Detail(user.GetVid())
	z.time = pro.Time
	z.mem = pro.Memory
	return nil
}

func (z *ZJGSUJudger) files(user vjudger.UserInterface, workdir string) {
	var codefilename string

	switch user.GetLang() {
	case config.LanguageC:
		codefilename = workdir + "/Main.c"
	case config.LanguageCPP:
		codefilename = workdir + "/Main.cc"
	case config.LanguageJAVA:
		codefilename = workdir + "/Main.java"
	case config.LanguagePY2:
		codefilename = workdir + "/Main.py2"
	case config.LanguagePY3:
		codefilename = workdir + "/Main.py3"
	}

	codefile, err := os.Create(codefilename)
	defer codefile.Close()

	_, err = codefile.WriteString(user.GetCode())
	if err != nil {
		logger.Println("source code writing to file failed")
	}
}

func (z *ZJGSUJudger) Submit(user vjudger.UserInterface) error {
	z.compile(user)

	if user.GetResult() != config.JudgeCE {
		user.SetResult(config.JudgeRJ)
		logger.Println("compile success")
		user.UpdateSolution()

		cmd := exec.Command("cp", "-r", dataPath+"/"+strconv.Itoa(user.GetVid()), runPath+"/"+strconv.Itoa(user.GetSid()))
		err := cmd.Run()
		if err != nil {
			log.Println("copy problem data failed")
			log.Println(err)
		}
	} else {
		b, err := ioutil.ReadFile(z.workdir + "/ce.txt")
		if err != nil {
			log.Println(err)
		}

		log.Println(string(b))
		user.SetErrorInfo(string(b))
		logger.Println("compiler error")
		log.Println("compiler error")

		return ErrCompile
	}
	return nil
}

func (z *ZJGSUJudger) GetStatus(user vjudger.UserInterface) error {
	logger.Println("run solution")

	var out bytes.Buffer
	cmd := exec.Command("runner", strconv.Itoa(user.GetLang()), strconv.Itoa(z.time), strconv.Itoa(z.mem), z.workdir)
	cmd.Stdout = &out
	cmd.Run()

	output := strings.Trim(out.String(), " ")
	logger.Printf("CJudger id: %v, output: %v\n", user.GetSid(), output)
	if len(output) == 0 {
		logger.Println("CJudger Result is wrong")
		user.SetResult(config.JudgeNA)
		user.SetResource(0, 0, len(user.GetCode()))
		return nil
	}

	sp := strings.Split(output, " ")
	if len(sp) != 3 {
		logger.Println("CJudger result split siz is wrong")
		user.SetResult(config.JudgeNA)
		user.SetResource(0, 0, len(user.GetCode()))
		return nil
	}
	var err error
	var Result, Time, Mem int
	Result, err = strconv.Atoi(sp[0])

	if err != nil {
		logger.Println(err)
		logger.Println(Result)
	}
	user.SetResult(Result)
	Time, err = strconv.Atoi(sp[1])
	Mem, err = strconv.Atoi(sp[2])
	Mem = Mem / 1024 //b->Kb
	user.SetResource(Time, Mem, len(user.GetCode()))
	return nil
}

func (z *ZJGSUJudger) Run(u vjudger.UserInterface) error {
	defer os.RemoveAll(runPath + "/" + strconv.Itoa(u.GetSid()))
	u.SetResult(config.JudgePD)
	for _, apply := range []func(vjudger.UserInterface) error{z.Init, z.Login, z.Submit, z.GetStatus} {
		if err := apply(u); err != nil {
			logger.Println(err)
			return err
		}
	}
	return nil
}

func (z *ZJGSUJudger) compile(user vjudger.UserInterface) {
	cmd := exec.Command("compiler", strconv.Itoa(user.GetLang()), z.workdir)
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	}

	if cmd.ProcessState.String() != "exit status 0" {
		log.Println(cmd.ProcessState.String())
		user.SetResult(config.JudgeCE) //compiler error
	}
}
