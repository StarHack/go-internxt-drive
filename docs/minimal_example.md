
Minimal example of how to use the client.

Run with `INTERNXT_TEST_EMAIL=<your email> INTERNXT_TEST_PASSWORD=<your password> go run main.go`


```go
package main

import (
	"fmt"
	"os"

	internxtclient "github.com/StarHack/go-internxt-drive/internxtclient"
)

func main() {
	email := os.Getenv("INTERNXT_TEST_EMAIL")
	password := os.Getenv("INTERNXT_TEST_PASSWORD")

	if email == "" || password == "" {
		fmt.Println("Please set the INTERNXT_TEST_EMAIL and INTERNXT_TEST_PASSWORD environment variables")
		return
	}

	c, err := internxtclient.NewWithCredentials(email, password)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	folders, err := c.Folders.Tree(c.UserData.AccessData.User.RootFolderUUID)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	println("Folders in rootfolder:")
	for _, folder := range folders.Children {
		fmt.Println(folder.PlainName)
	}

	println("\nFiles in rootfolder:")
	for _, file := range folders.Files {
		fmt.Println(file.PlainName + "." + file.Type)
	}
}
```