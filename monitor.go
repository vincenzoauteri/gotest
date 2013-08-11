package main

import (
  "os"
  "fmt"
  "time"
  "regexp"
)

var times map[string] int64

func parseDir (f *os.File) bool {
  result  := false

  dir, err := f.Readdir(100)
  if err != nil {
    return result
  }

  for i := range dir {
    if dir[i].Mode().IsDir() && dir[i].Name()[0] != '.'{
      os.Chdir(dir[i].Name())
      dirFile, err := os.Open(".")
      if (err == nil) {
        result = parseDir(dirFile)
        os.Chdir("..")
      } else {
        fmt.Println("Error")
      }
    } else {
      if matched, _ := regexp.MatchString(".go$",dir[i].Name()) ; matched { 
        path, _ := os.Getwd()
        key := path + "/" + dir[i].Name()
        value, none :=  times[key]
        currentModTime := dir[i].ModTime().Unix()
        if none {
          if currentModTime > value {
            fmt.Println("Modified: " + dir[i].Name())
            times[key] = currentModTime
            result = true
          }
        } else {
          times[key] = currentModTime
          result = true
        }
      }
    }
  }
  return result
}

func main() {
  var procAttr os.ProcAttr
  procAttr.Files = []*os.File{nil, os.Stdout, os.Stderr}
  args  := []string {"go","run","./wiki.go"}
  server ,err := os.StartProcess("/usr/local/bin/go",args,&procAttr)

  if err != nil {
    fmt.Println(err.Error())
  }
  times = make(map[string] int64)
  dir,err := os.Open(".")
  if err != nil {
    fmt.Println(err)
    return
  }
  _ = parseDir(dir)
  for {
    dir,_ := os.Open(".")
    mod := parseDir(dir)
    if mod {
      _ = server.Kill()
      server , _ = os.StartProcess("/usr/local/bin/go",args,&procAttr)
    }
    time.Sleep(1 * time.Second)
  }
}

