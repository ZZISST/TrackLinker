package s3

import (
	"bytes"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/pkg/config"
)

type Adapter struct {
	config config.Schema
}

func NewAdapter(config config.Schema) *Adapter {
	return &Adapter{
		config: config,
	}
}

func (a *Adapter) UploadTrackPreview(mp3URL, title, artist string) (previewAudioUrl *string, err error) {
	bucket := "track-previews"
	region := "ru-central1"
	endpoint := "https://storage.yandexcloud.net"
	accessKey := "REDACTED"
	secretKey := "REDACTED"

	fileKey := "tracks/previews/" + title + " " + artist + ".mp3"
	preview := "https://" + bucket + ".storage.yandexcloud.net/" + fileKey

	resp, err := http.Get(mp3URL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Read the response body into a buffer
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		return nil, err
	}

	// Configure the session for Yandex S3 (note: Yandex S3 requires S3ForcePathStyle)
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	uploader := s3manager.NewUploader(sess)

	// Upload the file; set ContentType to "audio/mpeg"
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fileKey),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("audio/mpeg"),
	})

	return &preview, err
}
