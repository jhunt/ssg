package main

import (
	"fmt"
	"io"
	"os"

	ssg "github.com/shieldproject/shield-storage-gateway/client"
)

func main() {
	ssgURL := os.Args[1]
	username := os.Args[2]
	password := os.Args[3]
	path := "/Users/srinikethvarma/go/src/github.com/jhunt/shield-storage-gateway/client/client/test.txt"

	control := ssg.NewControlClient(ssgURL, username, password)
	client := ssg.NewClient(ssgURL)

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

	fmt.Println("ID: ", upload.ID)
	fmt.Println("Token: ", upload.Token)

	download, err := control.StartDownload(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to start download: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("ID: ", download.ID)
	fmt.Println("Token: ", download.Token)

	in, err := client.Download(download.ID, download.Token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to attempt download: %s\n", err)
		os.Exit(1)
	}
	io.Copy(os.Stdout, in)
	in.Close()
}
