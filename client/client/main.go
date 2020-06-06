package main

import (
	"fmt"
	"io"
	"os"
	"time"

	ssg "github.com/jhunt/shield-storage-gateway/client"
)

func genBackupPath() string {
	t := time.Now()
	year, mon, day := t.Date()
	hour, min, sec := t.Clock()
	uuid := "238943-439834-34984-43934439"
	path := fmt.Sprintf("%04d-%02d-%02d-%02d%02d%02d-%s", year, mon, day, hour, min, sec, uuid)
	return path
}

func main() {
	ssgURL := os.Args[1]
	username := os.Args[2]
	password := os.Args[3]
	path := genBackupPath()

	fmt.Println(ssgURL, username, password, path)
	control := ssg.NewControlClient(ssgURL, username, password)
	client := ssg.NewClient(ssgURL)
	fmt.Println("Control: ", control.URL)

	upload, err := control.StartUpload(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to start upload: %s\n", err)
		os.Exit(1)
	}

	size, err := client.Upload(upload.ID, upload.Token, os.Stdin, true)
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

	fmt.Printf("\nSleep for 3 seconds...\n")
	time.Sleep(3 * time.Second)
	err = control.DeleteFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to start delete: %s\n", err)
		os.Exit(2)
	}
}
