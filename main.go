package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/webview/webview"
)

/*
#cgo darwin CXXFLAGS: -DWEBVIEW_COCOA -std=c++11
#cgo darwin LDFLAGS: -framework WebKit

#include <CoreGraphics/CoreGraphics.h>
#include <objc/objc-runtime.h>

extern void webview_set_fullscreen(void *m_window) {
	((void (*)(id, SEL, id))objc_msgSend)((id)m_window, sel_registerName("toggleFullScreen:"), (id)0);
}
*/
import "C"

func main() {
	// create an accessory
	info := accessory.Info{Name: "Fireplace"}
	ac := accessory.NewSwitch(info)
	w := webview.New(true)
	w.SetSize(800, 600, webview.HintNone)
	C.webview_set_fullscreen(w.Window())
	if dir, err := os.Getwd(); err != nil {
		panic(err)
	} else {
		w.Navigate("file://" + dir + "/index.html")
	}

	defer func() {
		if w != nil {
			w.Destroy()
		}
	}()

	ac.Switch.On.OnValueRemoteUpdate(func(on bool) {
		if on {
			if err := exec.Command("caffeinate", "-u", "-t", "1").Run(); err != nil {
				fmt.Println(err)
			}
			w.Dispatch(func() {
				w.Eval("video.play()")
			})
		} else {
			w.Dispatch(func() {
				w.Eval("video.pause()")
			})
			if err := exec.Command("pmset", "displaysleepnow").Run(); err != nil {
				fmt.Println(err)
			}
		}
		ac.Switch.On.SetValue(on)
	})

	ac.Switch.On.OnValueRemoteGet(func() bool {
		output, err := exec.Command("pmset", "-g", "powerstate", "IODisplayWrangler").CombinedOutput()
		if err != nil {
			fmt.Println(err)
		}
		return bytes.Contains(output, []byte("USEABLE"))
	})

	// configure the ip transport
	config := hc.Config{Pin: "00102003"}
	t, err := hc.NewIPTransport(config, ac.Accessory)
	if err != nil {
		log.Panic(err)
	}

	hc.OnTermination(func() {
		w.Terminate()
		<-t.Stop()
	})

	go t.Start()

	w.Run()
}
