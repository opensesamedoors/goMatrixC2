package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mholt/archiver"
	"github.com/wabarc/go-catbox"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

type PostMessage struct {
	MsgType string `json:"msgtype"`
	Body    string `json:"body"`
}

type GetMessage struct {
	Type    string `json:"type"`
	Content struct {
		Body string `json:"body"`
	} `json:"content"`
}

func ApiPost(body string) {
	client := &http.Client{}
	reqURL := fmt.Sprintf("%s/_matrix/client/r0/rooms/%s/send/m.room.message", HomeServer, RoomID)
	message := PostMessage{
		MsgType: "m.text",
		Body:    body,
	}
	messageJSON, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshaling message:", err)
		return
	}
	for {
		req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(messageJSON))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}
		req.Header.Add("Authorization", "Bearer "+AccessToken)
		req.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Response error: ", err)
			fmt.Println("Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()
		break
	}
}

func ApiGet() (string, error) {
	client := &http.Client{}
	reqURL := fmt.Sprintf("%s/_matrix/client/r0/rooms/%s/messages", HomeServer, url.PathEscape(RoomID))
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+AccessToken)
	q := req.URL.Query()
	q.Add("dir", "b")
	q.Add("limit", "1")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var result struct {
		Chunk []GetMessage `json:"chunk"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}
	if len(result.Chunk) > 0 {
		return result.Chunk[0].Content.Body, nil
	}
	return "", nil
}

func ApiUpload(filename string) {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}

	if fileInfo.IsDir() {
		archivePath := filename + ".zip"

		err := archiver.Archive([]string{filename}, archivePath)
		if err != nil {
			fmt.Println("Error archiving directory:", err)
			return
		}
		filename = archivePath
	}

	cat := catbox.New(nil)
	url, err := cat.Upload(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "catbox: %v\n", err)
		return
	}

	fmt.Fprintf(os.Stdout, "%s  %s\n", url, filename)

	ApiPost(fmt.Sprintf("Your file download link: %s", url))

	defer func() {
		time.Sleep(8 * time.Second)
		_ = os.Remove(filename)
	}()
}

func ApiDownload(link string, filepath string) error {
	resp, err := http.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fileInfo, err := os.Stat(filepath)
	if err == nil && fileInfo.IsDir() {
		filename := path.Base(link)
		filepath = path.Join(filepath, filename)
	} else {
		dir, file := path.Split(filepath)
		ext := path.Ext(file)
		name := strings.TrimSuffix(file, ext)
		newFilename := fmt.Sprintf("%s_RENAMED%s", name, ext)
		filepath = path.Join(dir, newFilename)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
