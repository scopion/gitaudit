package cmd

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"

	"git-audit/config"
	"git-audit/service"

	"github.com/serialt/sugar"
)

func env(key, def string) string {
	if x := os.Getenv(key); x != "" {
		return x
	}
	return def
}

var (
	appVersion bool
)

func init() {

	flag.BoolVarP(&appVersion, "version", "v", false, "Display build and version msg.")
	flag.StringVarP(&config.ConfigPath, "cfgFile", "c", env("CONFIG", config.ConfigPath), "Path to config yaml file.")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Println("使用说明")
		flag.PrintDefaults()
	}
	flag.ErrHelp = fmt.Errorf("\n\nSome errors have occurred, check and try again !!! ")

	flag.CommandLine.SortFlags = false
	flag.Parse()
	// register global var

}

func GitInit() {
	config.LoadConfig(config.ConfigPath)
	mylg := &sugar.Logger{
		LogLevel:      config.Config.GitLog.LogLevel,
		LogFile:       config.Config.GitLog.LogFile,
		LogType:       config.Config.GitLog.LogType,
		LogMaxSize:    50,
		LogMaxBackups: 3,
		LogMaxAge:     365,
		LogCompress:   true,
	}
	config.Logger = mylg.NewMyLogger()
	config.LogSugar = config.Logger.Sugar()
	service.LogSugar = config.Logger.Sugar()

	mydb := &sugar.Database{
		Type:     config.Config.Database.Type,
		Addr:     config.Config.Database.Addr,
		Port:     config.Config.Database.Port,
		DBName:   config.Config.Database.DBName,
		Username: config.Config.Database.Username,
		Password: config.Config.Database.Password,
	}
	config.DB = mydb.NewDBConnect(config.Logger)
}

func Run() {

	if appVersion {
		fmt.Printf("APPName: %v\n Maintainer: %v\n Version: %v\n BuildTime: %v\n GitCommit: %v\n GoVersion: %v\n OS/Arch: %v\n",
			config.APPName,
			config.Maintainer,
			config.APPVersion,
			config.BuildTime,
			config.GitCommit,
			config.GOVERSION,
			config.GOOSARCH)
		return
	}
	GitInit()

	// pkg.Sugar.Info(config.LogFile)
	config.WG.Add(2)
	go func() {
		service.SSHService()
		config.WG.Done()
	}()
	go func() {
		service.HTTPService()
		config.WG.Done()
	}()
	config.WG.Wait()
	os.Exit(0)
	return
}
