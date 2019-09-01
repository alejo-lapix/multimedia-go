package service

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"runtime"
	"testing"

	"github.com/alejo-lapix/multimedia-go/persistence"

	"github.com/aws/aws-sdk-go/aws"
)

var _, filePath, _, _ = runtime.Caller(1)

func TestHttpFileUploader_MoveFile(t *testing.T) {
	type fields struct {
		Uploader  Uploader
		MaxUpload int64
	}
	type args struct {
		request *http.Request
		key     string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		/**
		TODO Test with a really big file
		{
			name:   "If file is bigger that permitted",
			fields: fields{MaxUpload: 0},
			args: args{
				request: newMultipartRequest("file", filePath),
				key:     "file",
			},
			wantErr: true,
		},
		*/
		{
			name:   "If file key is empty returns an error",
			fields: fields{MaxUpload: 5},
			args: args{
				request: newMultipartRequest("wrong-key", filePath),
				key:     "file",
			},
			wantErr: true,
		},
		{
			name:   "If uploader return error it return error",
			fields: fields{MaxUpload: 5, Uploader: &FailUploader{}},
			args: args{
				request: newMultipartRequest("file", filePath),
				key:     "file",
			},
			wantErr: true,
		},
		{
			name:   "Should return a MultimediaItem",
			fields: fields{MaxUpload: 5, Uploader: &SuccessUploader{}},
			args: args{
				request: newMultipartRequest("file", filePath),
				key:     "file",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uploader := &HttpFileUploader{
				Uploader:      tt.fields.Uploader,
				MaxMBUploaded: tt.fields.MaxUpload,
			}
			got, err := uploader.MoveFile(tt.args.request, &tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("MoveFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("MoveFile() returns nil")
			}
		})
	}
}

type UploadFileError struct{}

func (err UploadFileError) Error() string {
	return "Error while trying to upload a file"
}

type FailUploader struct{}

func (uploader *FailUploader) Upload(filename *string, destination *string) (*persistence.MultimediaItem, error) {
	return nil, UploadFileError{}
}

func (uploader *FailUploader) Delete(ID *string) error {
	return UploadFileError{}
}

type SuccessUploader struct{}

func (uploader *SuccessUploader) Upload(filename *string, destination *string) (*persistence.MultimediaItem, error) {
	return persistence.NewMultimediaItem(
		aws.String("http://any-bucket.dev"),
		aws.String("file-name"),
		aws.String(persistence.IMAGE),
	)
}

func (uploader *SuccessUploader) Delete(ID *string) error {
	return nil
}

func newMultipartRequest(paramName, path string) *http.Request {
	req, err := newFileUploadRequest("http://localhost:8080", map[string]string{}, paramName, path)

	if err != nil {
		panic(err)
	}

	return req
}

// Creates a new file upload http request with optional extra params
func newFileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, fi.Name())
	if err != nil {
		return nil, err
	}
	_, err = part.Write(fileContents)
	if err != nil {
		return nil, err
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	return req, nil
}

func TestIOFileUploader_MoveFile(t *testing.T) {
	thisFileReader, name, size := thisFileIOReader()
	type fields struct {
		Uploader Uploader
	}
	type args struct {
		ioReader io.Reader
		fileName string
		fileSize int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *persistence.MultimediaItem
		wantErr bool
	}{
		{
			name: "Uploads a file from IOFileUploader",
			fields: fields{
				Uploader: &SuccessUploader{},
			},
			args: args{
				ioReader: thisFileReader,
				fileName: name,
				fileSize: size,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uploader := &IOFileUploader{
				Uploader: tt.fields.Uploader,
			}
			got, err := uploader.MoveFile(tt.args.ioReader, tt.args.fileName, tt.args.fileSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("MoveFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("MoveFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func thisFileIOReader() (file io.Reader, filename string, size int64) {
	ioReader, err := os.Open(filePath)

	if err != nil {
		panic(err)
	}

	stats, err := os.Stat(filePath)

	if err != nil {
		panic(err)
	}

	return ioReader, "http_test.go", stats.Size()
}
