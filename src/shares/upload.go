package shares

import (
	"fmt"
	"time"
	"../fs"
	"../config"
	"../cloud"
)

var upload_delay = 5 * time.Second

func (ss *Share) upload() {
	ss.Uploads <-true
}

func (ss *Share) uploadLoop() {
	delay := time.NewTimer(upload_delay)

	for {
		select {
		case again := <-ss.Uploads:
			if again {
				delay.Reset(upload_delay)
			} else {
				fmt.Println("Shutting down uploadLoop")
				break
			}
		case _ = <-delay.C:
			ss.reallyUpload()
		}
	}
}

func (ss *Share) reallyUpload() {
	settings := config.GetSettings()
	if !settings.Ready() {
		fmt.Println("Skipping upload, no cloud configured.")
		return
	}

	// Check Cloud Share Setup
	cc, err := cloud.New()
	if err != nil {
		fmt.Println(fs.Trace(err))
		return
	}

	sdata, err := cc.GetShare(ss.NameHmac())
	if err == cloud.ErrNotFound {
		fmt.Println("XX - Creating share")
		sdata, err = cc.CreateShare(ss.NameHmac())
		fmt.Println("XX - Created")
	} 
	if err != nil {
		panic(err)
	}
	fmt.Println(sdata)

	// TODO: Verify that this is a fast-forward
	

	// Upload
	cp, err := ss.Trie.MakeCheckpoint()
	fs.CheckError(err)
	defer cp.Cleanup()

	ba, err := ss.Trie.NewArchive()
	if err != nil {
		panic(err)
	}
	defer ba.Close()

	err = ba.AddList(cp.Adds)
	if err != nil {
		panic(err)
	}

	err = cc.SendBlocks(ss.NameHmac(), ba.FileName())
	if err != nil {
		panic(err)
	}

	/*
	fmt.Println("== Added Blocks ==")
	addsData := pio.ReadFile(cp.Adds)
	fmt.Println(string(addsData))

	fmt.Println("== Dead Blocks ==")
	deadData := pio.ReadFile(cp.Dels)
	fmt.Println(string(deadData))
	*/
}