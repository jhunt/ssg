package main

import (
	"fmt"
	"io"
	"os"
	"time"

	ssg "github.com/shieldproject/shield-storage-gateway/client"
)

func main() {
	ssgURL := os.Args[1]
	username := os.Args[2]
	password := os.Args[3]
	path := "/Users/srinikethvarma/go/src/github.com/jhunt/shield-storage-gateway/client/client/test.txt"

	fmt.Println(ssgURL, username, password, path)
	control := ssg.NewControlClient(ssgURL, username, password)
	client := ssg.NewClient(ssgURL)
	fmt.Println("Control: ", control.URL)

	upload, err := control.StartUpload(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to start upload: %s\n", err)
		os.Exit(1)
	}

	size, err := client.Upload(upload.ID, upload.Token, os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to upload to the storage gateway: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Size: ", size)

	download, err := control.StartDownload(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to start download: %s\n", err)
		os.Exit(1)
	}

	in, err := client.Download(download.ID, download.Token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to attempt download: %s\n", err)
		os.Exit(1)
	}
	io.Copy(os.Stdout, in)
	in.Close()

	delete, err := control.StartDelete(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to start delete: %s\n", err)
		os.Exit(2)
	}

	fmt.Printf("Delete ID:  %s\n", delete.ID)
	fmt.Printf("Delete Token: %s\n", delete.Token)
	fmt.Printf("Sleeping for 3 seconds...\n")
	time.Sleep(3 * time.Second)

	err = client.Delete(delete.ID, path, delete.Token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "usable to start delete: %s \n", err)
		os.Exit(2)
	}
}
