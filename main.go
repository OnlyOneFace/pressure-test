package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"pressure-test/config"
	"pressure-test/handle"
)

func initConfig() {
	num := runtime.NumCPU()
	fmt.Printf("go version:%v cpu num:%v\n", runtime.Version(), num)
	runtime.GOMAXPROCS(num) //开启电脑最大数量CPU
	flag.BoolVar(&config.Cfg.DebugMode, "debug", false, "mode switch")
	flag.StringVar(&config.Cfg.SoFilePath, "sp", "", ".so file path")
	flag.StringVar(&config.Cfg.CaseName, "caseName", "test", "case name")
	flag.StringVar(&config.Cfg.FileStorage, "f", "file/", "memery profile")
	flag.StringVar(&config.Cfg.Cpuprofile, "cpuprofile", "cpu.prof", "cpu profile")
	flag.StringVar(&config.Cfg.Memprofile, "memprofile", "mem.prof", "memery profile")
	// 解析参数
	flag.Parse()
	if _, err := os.Stat(config.Cfg.FileStorage); os.IsNotExist(err) {
		if err = os.MkdirAll(config.Cfg.FileStorage, os.ModeDir); err != nil {
			log.Fatal("init make directory: ", err)
		}
	}
}

func main() {
	initConfig()
	//cpu信息
	cf, err := os.Create(config.Cfg.FileStorage + config.Cfg.Cpuprofile)
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer cf.Close()
	if err = pprof.StartCPUProfile(cf); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()
	//逻辑
	handle.StartBenchmark()
	//内存堆栈信息
	var mf *os.File
	if mf, err = os.Create(config.Cfg.FileStorage + config.Cfg.Memprofile); err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer mf.Close()
	runtime.GC()
	if err = pprof.WriteHeapProfile(mf); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
}
