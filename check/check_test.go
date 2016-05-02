package check_test

import (
	"crypto/rand"
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/ljfranklin/terraform-resource/check"
	"github.com/ljfranklin/terraform-resource/in/models"
	baseModels "github.com/ljfranklin/terraform-resource/models"
	"github.com/ljfranklin/terraform-resource/storage"
	"github.com/ljfranklin/terraform-resource/test/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check", func() {

	var (
		checkInput          models.InRequest
		bucket              string
		prevEnvName         string
		currEnvName         string
		pathToPrevS3Fixture string
		pathToCurrS3Fixture string
		awsVerifier         *helpers.AWSVerifier
	)

	BeforeEach(func() {
		accessKey := os.Getenv("AWS_ACCESS_KEY")
		Expect(accessKey).ToNot(BeEmpty(), "AWS_ACCESS_KEY must be set")

		secretKey := os.Getenv("AWS_SECRET_KEY")
		Expect(secretKey).ToNot(BeEmpty(), "AWS_SECRET_KEY must be set")

		bucket = os.Getenv("AWS_BUCKET")
		Expect(bucket).ToNot(BeEmpty(), "AWS_BUCKET must be set")

		bucketPath := os.Getenv("AWS_BUCKET_PATH")
		Expect(bucketPath).ToNot(BeEmpty(), "AWS_BUCKET_PATH must be set")

		region := os.Getenv("AWS_REGION") // optional
		if region == "" {
			region = "us-east-1"
		}

		awsVerifier = helpers.NewAWSVerifier(
			accessKey,
			secretKey,
			region,
		)
		prevEnvName = randomString("s3-test-fixture-previous")
		pathToPrevS3Fixture = path.Join(bucketPath, prevEnvName)
		currEnvName = randomString("s3-test-fixture-current")
		pathToCurrS3Fixture = path.Join(bucketPath, currEnvName)

		checkInput = models.InRequest{
			Source: models.Source{
				Storage: storage.Model{
					Bucket:          bucket,
					BucketPath:      bucketPath,
					AccessKeyID:     accessKey,
					SecretAccessKey: secretKey,
				},
			},
		}
	})

	AfterEach(func() {
		awsVerifier.DeleteObjectFromS3(bucket, pathToPrevS3Fixture)
		awsVerifier.DeleteObjectFromS3(bucket, pathToCurrS3Fixture)
	})

	Context("when bucket is empty", func() {
		It("returns an empty version list", func() {
			runner := check.Runner{}
			resp, err := runner.Run(checkInput)
			Expect(err).ToNot(HaveOccurred())

			expectedOutput := []baseModels.Version{}
			Expect(resp).To(Equal(expectedOutput))
		})
	})

	Context("when bucket contains multiple state files", func() {
		BeforeEach(func() {
			prevFixture, err := os.Open(getFileLocation("fixtures/s3/terraform-previous.tfstate"))
			Expect(err).ToNot(HaveOccurred())
			defer prevFixture.Close()

			awsVerifier.UploadObjectToS3(bucket, pathToPrevS3Fixture, prevFixture)
			time.Sleep(5 * time.Second) // ensure last modified is different

			currFixture, err := os.Open(getFileLocation("fixtures/s3/terraform-current.tfstate"))
			Expect(err).ToNot(HaveOccurred())
			defer currFixture.Close()
			awsVerifier.UploadObjectToS3(bucket, pathToCurrS3Fixture, currFixture)
		})

		It("returns the latest version on S3", func() {
			runner := check.Runner{}
			resp, err := runner.Run(checkInput)
			Expect(err).ToNot(HaveOccurred())

			lastModified := awsVerifier.GetLastModifiedFromS3(bucket, pathToCurrS3Fixture)

			expectOutput := []baseModels.Version{
				baseModels.Version{
					LastModified: lastModified,
					EnvName:      currEnvName,
				},
			}
			Expect(resp).To(Equal(expectOutput))
		})

		It("returns an empty version list when current version matches storage version", func() {
			currentLastModified := awsVerifier.GetLastModifiedFromS3(bucket, pathToCurrS3Fixture)
			checkInput.Version = baseModels.Version{
				LastModified: currentLastModified,
				EnvName:      currEnvName,
			}

			runner := check.Runner{}
			resp, err := runner.Run(checkInput)
			Expect(err).ToNot(HaveOccurred())

			expectOutput := []baseModels.Version{}
			Expect(resp).To(Equal(expectOutput))
		})
	})
})

func randomString(prefix string) string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	Expect(err).ToNot(HaveOccurred())
	return fmt.Sprintf("%s-%x", prefix, b)
}

func getFileLocation(relativePath string) string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Join(path.Dir(filename), "..", relativePath)
}
