package service

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/alejo-lapix/multimedia-go/persistence"

	"github.com/google/uuid"
)

type HttpFileUploader struct {
	Uploader      Uploader
	MaxMBUploaded int64
}

type IOFileUploader struct {
	Uploader Uploader
}

func (uploader *IOFileUploader) MoveFile(ioReader io.Reader, fileName string, fileSize int64) (*persistence.MultimediaItem, error) {
	buffer := make([]byte, fileSize)
	fileExtension := path.Ext(fileName)
	temporalFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("upload-*%v", fileExtension))

	if err != nil {
		return nil, err
	}

	defer temporalFile.Close()

	_, err = ioReader.Read(buffer)

	if err != nil {
		return nil, err
	}

	_, err = temporalFile.Write(buffer)

	if err != nil {
		return nil, err
	}

	ID := uuid.New().ID()
	filePath := temporalFile.Name()
	newFileName := fmt.Sprintf("%v-%v%v", time.Now().Format("20060102150405"), ID, fileExtension)

	item, err := uploader.Uploader.Upload(&filePath, &newFileName)

	if err != nil {
		return nil, err
	}

	return item, nil
}

// MoveFile moves a file to the given Uploader configuration
func (uploader *HttpFileUploader) MoveFile(request *http.Request, key *string) (*persistence.MultimediaItem, error) {
	var fileExtension string

	err := request.ParseMultipartForm(uploader.MaxMBUploaded << 20)

	if err != nil {
		return nil, err
	}

	file, handler, err := request.FormFile(*key)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	fileExtension = path.Ext(handler.Filename)
	temporalFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("upload-*%v", fileExtension))

	if err != nil {
		return nil, err
	}

	defer temporalFile.Close()

	fileBytes, err := ioutil.ReadAll(file)

	if err != nil {
		return nil, err
	}

	_, err = temporalFile.Write(fileBytes)

	if err != nil {
		return nil, err
	}

	ID := uuid.New().ID()
	filePath := temporalFile.Name()
	fileName := fmt.Sprintf("%v-%v.%v", time.Now().Format("20060102150405"), ID, fileExtension)

	item, err := uploader.Uploader.Upload(&filePath, &fileName)

	if err != nil {
		return nil, err
	}

	return item, nil
}
