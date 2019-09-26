package main

import (
    "fmt"
    "os/signal"
    "syscall"
    "time"
    "os"
    "net"

    "github.com/milosgajdos83/tenus"
    "github.com/lightstar-dev/openlan-go/libol"
    "github.com/lightstar-dev/openlan-go/point"
)

func UpLink(name string, c *point.Config) error {
    libol.Debug("main.UpLink: %s", name)
    link, err := tenus.NewLinkFrom(name)
    if err != nil {
        libol.Error("main.UpLink: Get ifce %s: %s", name, err)
        return err
    }
    
    if err := link.SetLinkUp(); err != nil {
        libol.Error("main.UpLink.SetLinkUp: %s : %s", name, err)
        return err
    }
    
    ip, ipnet, err := net.ParseCIDR(c.Ifaddr)
    if err != nil {
        libol.Error("main.UpLink.ParseCIDR %s : %s", c.Ifaddr, err)
        return err
    }

    if c.Brname != "" {
        br, err := tenus.BridgeFromName(c.Brname)
        if err != nil {
            libol.Error("main.UpLink.newBr: %s", err)
            br, err = tenus.NewBridgeWithName(c.Brname)
            if err != nil {
                libol.Error("main.UpLink.newBr: %s", err)
            }
        }

        if err := br.SetLinkUp(); err != nil {
            libol.Error("main.UpLink.newBr.Up: %s", err)
        }

        if err := br.AddSlaveIfc(link.NetInterface()); err != nil {
            libol.Error("main.UpLink.AddSlave: Switch ifce %s: %s", name, err)
        }

        link, err = tenus.NewLinkFrom(c.Brname)
        if err != nil {
            libol.Error("main.UpLink: Get ifce %s: %s", c.Brname, err)
        }
    }
    
    if err := link.SetLinkIp(ip, ipnet); err != nil {
        libol.Error("main.UpLink.SetLinkIp : %s", err)
        return err
    }
	
    return nil
}

func main() {
    c := point.NewConfig()
    libol.Debug("main.config: %s", c)

    p := point.NewPoint(c)
    
    UpLink(p.Ifce.Name(), c)
    p.Start()

    x := make(chan os.Signal)
    signal.Notify(x, os.Interrupt, syscall.SIGTERM)
    go func() {
        <- x
        p.Close()
        fmt.Println("Done!")
        os.Exit(0)
    }()

    fmt.Println("Please enter CTRL+C to exit...")
    for {
        time.Sleep(1000 * time.Second)
    }
}
