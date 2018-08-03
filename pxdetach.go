//
// fidn and detach all locally attachd but not mounted portworx volumes
//
package main

import (
    "bufio"
    "io/ioutil"
    "log"
    "os"
    "strings"
    "net/http"
)

func getpxmounts() map[string]bool {
    m := make(map[string]bool)

    pids, err := ioutil.ReadDir("/proc")
    if err != nil {
        log.Fatal(err)
    }

    for _, pid := range pids {
        if pid.IsDir() && pid.Name()[0] >= '1' && pid.Name()[0] <= '9' {
            mounts, err := os.Open("/proc/" + pid.Name() + "/mounts")
            if err == nil {
                scanner := bufio.NewScanner(mounts)
                scanner.Split(bufio.ScanLines)
                for scanner.Scan() {
                    f := strings.Fields(scanner.Text())
                    if strings.HasPrefix(f[0], "/dev/pxd/pxd") {
                        m[f[0]]=true
                    }
                }
                mounts.Close()
            }
        }
    }

    return m
}

func getpxattach() map[string]bool {
    a := make(map[string]bool)

    pxd, err := ioutil.ReadDir("/dev/pxd")
    if err != nil {
        log.Fatal(err)
    }

    for _, dev := range pxd {
        if !strings.HasPrefix(dev.Name(), "pxd-control") {
            a["/dev/pxd/" + dev.Name()]=true
        }
    }

    return a
}

func detach(vol string) {
    log.Print("Detaching Volume ", vol)
    h := &http.Client{}
    req, err := http.NewRequest("PUT", "http://127.0.0.1:9001/v1/osd-volumes/" + vol, strings.NewReader("{\"action\":{\"attach\":1},\"options\":{\"UNMOUNT_BEFORE_DETACH\":\"false\"}}"))
    resp, err := h.Do(req)
    if err != nil {
        log.Print("Error detaching ", vol, " ", err)
        return
    }
    log.Print(resp.Status)
    resp.Body.Close()
}


func main() {
    attach:=getpxattach()
    mounts:=getpxmounts()

    for a, _ := range attach {
        for m, _ := range mounts {
            if a == m {
                attach[a]=false
            }
        }
    }

    for a, v := range attach {
        if v {
            detach(strings.TrimPrefix(a, "/dev/pxd/pxd"))
        }
    }
}
