//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package digitaloceanimport

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/digitalocean/godo"

	"github.com/digitalocean/packer-plugin-digitalocean/builder/digitalocean"
	"github.com/digitalocean/packer-plugin-digitalocean/version"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/useragent"
)

const BuilderId = "packer.post-processor.digitalocean-import"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// A personal access token used to communicate with the DigitalOcean v2 API.
	// This may also be set using the `DIGITALOCEAN_TOKEN` or
	// `DIGITALOCEAN_ACCESS_TOKEN` environmental variables.
	APIToken string `mapstructure:"api_token" required:"true"`
	// The access key used to communicate with Spaces. This may also be set using
	// the `DIGITALOCEAN_SPACES_ACCESS_KEY` environmental variable.
	SpacesKey string `mapstructure:"spaces_key" required:"true"`
	// The secret key used to communicate with Spaces. This may also be set using
	// the `DIGITALOCEAN_SPACES_SECRET_KEY` environmental variable.
	SpacesSecret string `mapstructure:"spaces_secret" required:"true"`
	// The maximum number of retries for requests that fail with a 429 or 500-level error.
	// The default value is 5. Set to 0 to disable reties.
	HTTPRetryMax *int `mapstructure:"http_retry_max" required:"false"`
	// The maximum wait time (in seconds) between failed API requests. Default: 30.0
	HTTPRetryWaitMax *float64 `mapstructure:"http_retry_wait_max" required:"false"`
	// The minimum wait time (in seconds) between failed API requests. Default: 1.0
	HTTPRetryWaitMin *float64 `mapstructure:"http_retry_wait_min" required:"false"`
	// The name of the region, such as `nyc3`, in which to upload the image to Spaces.
	SpacesRegion string `mapstructure:"spaces_region" required:"true"`
	// The name of the specific Space where the image file will be copied to for
	// import. This Space must exist when the post-processor is run.
	SpaceName string `mapstructure:"space_name" required:"true"`
	// The name of the key used in the Space where the image file will be copied
	// to for import. This is treated as a [template engine](/docs/templates/legacy_json_templates/engine).
	// Therefore, you may use user variables and template functions in this field.
	// If not specified, this will default to `packer-import-{{timestamp}}`.
	ObjectName string `mapstructure:"space_object_name"`
	// Whether we should skip removing the image file uploaded to Spaces after
	// the import process has completed. "true" means that we should leave it in
	// the Space, "false" means to clean it out. Defaults to `false`.
	SkipClean bool `mapstructure:"skip_clean"`
	// A list of tags to apply to the resulting imported image.
	Tags []string `mapstructure:"image_tags"`
	// The name to be used for the resulting DigitalOcean custom image.
	Name string `mapstructure:"image_name" required:"true"`
	// The description to set for the resulting imported image.
	Description string `mapstructure:"image_description"`
	// The name of the distribution to set for the resulting imported image.
	Distribution string `mapstructure:"image_distribution"`
	// A list of DigitalOcean regions, such as `nyc3`, where the resulting image
	// will be available for use in creating Droplets.
	ImageRegions []string `mapstructure:"image_regions" required:"true"`
	// The length of time in minutes to wait for individual steps in the process
	// to successfully complete. This includes both importing the image from Spaces
	// as well as distributing the resulting image to additional regions. If not
	// specified, this will default to 20.
	Timeout time.Duration `mapstructure:"timeout"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
}

type apiTokenSource struct {
	AccessToken string
}

type logger struct {
	logger *log.Logger
}

func (t *apiTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: t.AccessToken,
	}, nil
}

func (l logger) Log(args ...interface{}) {
	l.logger.Println(args...)
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{"space_object_name"},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if p.config.SpacesKey == "" {
		p.config.SpacesKey = os.Getenv("DIGITALOCEAN_SPACES_ACCESS_KEY")
	}

	if p.config.SpacesSecret == "" {
		p.config.SpacesSecret = os.Getenv("DIGITALOCEAN_SPACES_SECRET_KEY")
	}

	if p.config.APIToken == "" {
		p.config.APIToken = os.Getenv("DIGITALOCEAN_TOKEN")
	}
	if p.config.APIToken == "" {
		p.config.APIToken = os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	}
	if p.config.APIToken == "" {
		p.config.APIToken = os.Getenv("DIGITALOCEAN_API_TOKEN")
	}

	if p.config.HTTPRetryMax == nil {
		p.config.HTTPRetryMax = godo.PtrTo(5)
		if max := os.Getenv("DIGITALOCEAN_HTTP_RETRY_MAX"); max != "" {
			maxInt, err := strconv.Atoi(max)
			if err != nil {
				return err
			}
			p.config.HTTPRetryMax = godo.PtrTo(maxInt)
		}
	}
	if p.config.HTTPRetryWaitMax == nil {
		p.config.HTTPRetryWaitMax = godo.PtrTo(30.0)
		if waitMax := os.Getenv("DIGITALOCEAN_HTTP_RETRY_WAIT_MAX"); waitMax != "" {
			waitMaxFloat, err := strconv.ParseFloat(waitMax, 64)
			if err != nil {
				return err
			}
			p.config.HTTPRetryWaitMax = godo.PtrTo(waitMaxFloat)
		}
	}
	if p.config.HTTPRetryWaitMin == nil {
		p.config.HTTPRetryWaitMin = godo.PtrTo(1.0)
		if waitMin := os.Getenv("DIGITALOCEAN_HTTP_RETRY_WAIT_MIN"); waitMin != "" {
			waitMinFloat, err := strconv.ParseFloat(waitMin, 64)
			if err != nil {
				return err
			}
			p.config.HTTPRetryWaitMin = godo.PtrTo(waitMinFloat)
		}
	}

	if p.config.ObjectName == "" {
		p.config.ObjectName = "packer-import-{{timestamp}}"
	}

	if p.config.Distribution == "" {
		p.config.Distribution = "Unkown"
	}

	if p.config.Timeout == 0 {
		p.config.Timeout = 20 * time.Minute
	}

	errs := new(packersdk.MultiError)

	if err = interpolate.Validate(p.config.ObjectName, &p.config.ctx); err != nil {
		errs = packersdk.MultiErrorAppend(
			errs, fmt.Errorf("Error parsing space_object_name template: %s", err))
	}

	requiredArgs := map[string]*string{
		"api_token":     &p.config.APIToken,
		"spaces_key":    &p.config.SpacesKey,
		"spaces_secret": &p.config.SpacesSecret,
		"spaces_region": &p.config.SpacesRegion,
		"space_name":    &p.config.SpaceName,
		"image_name":    &p.config.Name,
	}
	for key, ptr := range requiredArgs {
		if *ptr == "" {
			errs = packersdk.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
		}
	}

	if len(p.config.ImageRegions) == 0 {
		errs = packersdk.MultiErrorAppend(
			errs, fmt.Errorf("image_regions must be set"))
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	packersdk.LogSecretFilter.Set(p.config.SpacesKey, p.config.SpacesSecret, p.config.APIToken)
	log.Println(p.config)
	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, artifact packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
	var err error

	generatedData := artifact.State("generated_data")
	if generatedData == nil {
		// Make sure it's not a nil map so we can assign to it later.
		generatedData = make(map[string]interface{})
	}
	p.config.ctx.Data = generatedData

	p.config.ObjectName, err = interpolate.Render(p.config.ObjectName, &p.config.ctx)
	if err != nil {
		return nil, false, false, fmt.Errorf("Error rendering space_object_name template: %s", err)
	}
	log.Printf("Rendered space_object_name as %s", p.config.ObjectName)

	log.Println("Looking for image in artifact")
	source, err := extractImageArtifact(artifact.Files())
	if err != nil {
		return nil, false, false, fmt.Errorf("Image file not found")
	}

	spacesCreds := credentials.NewStaticCredentials(p.config.SpacesKey, p.config.SpacesSecret, "")
	spacesEndpoint := fmt.Sprintf("https://%s.digitaloceanspaces.com", p.config.SpacesRegion)
	spacesConfig := &aws.Config{
		Credentials: spacesCreds,
		Endpoint:    aws.String(spacesEndpoint),
		Region:      aws.String(p.config.SpacesRegion),
		LogLevel:    aws.LogLevel(aws.LogDebugWithSigning),
		Logger: &logger{
			logger: log.New(os.Stderr, "", log.LstdFlags),
		},
	}
	sess, err := session.NewSession(spacesConfig)
	if err != nil {
		return nil, false, false, err
	}

	ui.Message(fmt.Sprintf("Uploading %s to spaces://%s/%s", source, p.config.SpaceName, p.config.ObjectName))
	err = uploadImageToSpaces(source, p, sess)
	if err != nil {
		return nil, false, false, err
	}
	ui.Message(fmt.Sprintf("Completed upload of %s to spaces://%s/%s", source, p.config.SpaceName, p.config.ObjectName))

	ua := useragent.String(version.PluginVersion.FormattedVersion())
	opts := []godo.ClientOpt{godo.SetUserAgent(ua)}

	if *p.config.HTTPRetryMax > 0 {
		opts = append(opts, godo.WithRetryAndBackoffs(godo.RetryConfig{
			RetryMax:     *p.config.HTTPRetryMax,
			RetryWaitMin: p.config.HTTPRetryWaitMin,
			RetryWaitMax: p.config.HTTPRetryWaitMax,
			Logger:       log.Default(),
		}))
	}

	client, err := godo.New(oauth2.NewClient(context.TODO(), &apiTokenSource{
		AccessToken: p.config.APIToken,
	}), opts...)
	if err != nil {
		return nil, false, false, fmt.Errorf("DigitalOcean: could not create client, %s", err)
	}

	ui.Message(fmt.Sprintf("Started import of spaces://%s/%s", p.config.SpaceName, p.config.ObjectName))
	image, err := importImageFromSpaces(p, client)
	if err != nil {
		return nil, false, false, err
	}

	ui.Message(fmt.Sprintf("Waiting for import of image %s to complete (may take a while)", p.config.Name))
	err = waitUntilImageAvailable(client, image.ID, p.config.Timeout)
	if err != nil {
		return nil, false, false, fmt.Errorf("Import of image %s failed with error: %s", p.config.Name, err)
	}
	ui.Message(fmt.Sprintf("Import of image %s complete", p.config.Name))

	if len(p.config.ImageRegions) > 1 {
		// Remove the first region from the slice as the image is already there.
		regions := p.config.ImageRegions
		regions[0] = regions[len(regions)-1]
		regions[len(regions)-1] = ""
		regions = regions[:len(regions)-1]

		ui.Message(fmt.Sprintf("Distributing image %s to additional regions: %v", p.config.Name, regions))
		err = distributeImageToRegions(client, image.ID, regions, p.config.Timeout)
		if err != nil {
			return nil, false, false, err
		}
	}

	log.Printf("Adding created image ID %v to output artifacts", image.ID)
	artifact = &digitalocean.Artifact{
		SnapshotName: image.Name,
		SnapshotId:   image.ID,
		RegionNames:  p.config.ImageRegions,
		Client:       client,
	}

	if !p.config.SkipClean {
		ui.Message(fmt.Sprintf("Deleting import source spaces://%s/%s", p.config.SpaceName, p.config.ObjectName))
		err = deleteImageFromSpaces(p, sess)
		if err != nil {
			return nil, false, false, err
		}
	}

	return artifact, false, false, nil
}

func extractImageArtifact(artifacts []string) (string, error) {
	artifactCount := len(artifacts)

	if artifactCount == 0 {
		return "", fmt.Errorf("no artifacts were provided")
	}

	if artifactCount == 1 {
		return artifacts[0], nil
	}

	validSuffix := []string{"raw", "img", "qcow2", "vhdx", "vdi", "vmdk", "tar.bz2", "tar.xz", "tar.gz"}
	for _, path := range artifacts {
		for _, suffix := range validSuffix {
			if strings.HasSuffix(path, suffix) {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("no valid image file found")
}

func uploadImageToSpaces(source string, p *PostProcessor, s *session.Session) (err error) {
	file, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("Failed to open %s: %s", source, err)
	}

	uploader := s3manager.NewUploader(s)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Body:   file,
		Bucket: &p.config.SpaceName,
		Key:    &p.config.ObjectName,
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return fmt.Errorf("Failed to upload %s: %s", source, err)
	}

	file.Close()

	return nil
}

func importImageFromSpaces(p *PostProcessor, client *godo.Client) (image *godo.Image, err error) {
	log.Printf("Importing custom image from spaces://%s/%s", p.config.SpaceName, p.config.ObjectName)

	url := fmt.Sprintf("https://%s.%s.digitaloceanspaces.com/%s", p.config.SpaceName, p.config.SpacesRegion, p.config.ObjectName)
	createRequest := &godo.CustomImageCreateRequest{
		Name:         p.config.Name,
		Url:          url,
		Region:       p.config.ImageRegions[0],
		Distribution: p.config.Distribution,
		Description:  p.config.Description,
		Tags:         p.config.Tags,
	}

	image, _, err = client.Images.Create(context.TODO(), createRequest)
	if err != nil {
		return image, fmt.Errorf("Failed to import from spaces://%s/%s: %s", p.config.SpaceName, p.config.ObjectName, err)
	}

	return image, nil
}

func waitUntilImageAvailable(client *godo.Client, imageId int, timeout time.Duration) (err error) {
	done := make(chan struct{})
	defer close(done)

	result := make(chan error, 1)
	go func() {
		attempts := 0
		for {
			attempts += 1

			log.Printf("Waiting for image to become available... (attempt: %d)", attempts)
			image, _, err := client.Images.GetByID(context.TODO(), imageId)
			if err != nil {
				result <- err
				return
			}

			if image.Status == "available" {
				result <- nil
				return
			}

			if image.ErrorMessage != "" {
				result <- fmt.Errorf("%v", image.ErrorMessage)
				return
			}

			time.Sleep(3 * time.Second)

			select {
			case <-done:
				return
			default:
			}
		}
	}()

	log.Printf("Waiting for up to %d seconds for image to become available", timeout/time.Second)
	select {
	case err := <-result:
		return err
	case <-time.After(timeout):
		err := fmt.Errorf("Timeout while waiting to for action to become available")
		return err
	}
}

func distributeImageToRegions(client *godo.Client, imageId int, regions []string, timeout time.Duration) (err error) {
	for _, region := range regions {
		transferRequest := &godo.ActionRequest{
			"type":   "transfer",
			"region": region,
		}
		log.Printf("Transferring image to %s", region)
		action, _, err := client.ImageActions.Transfer(context.TODO(), imageId, transferRequest)
		if err != nil {
			return fmt.Errorf("Error transferring image: %s", err)
		}

		if err := digitalocean.WaitForImageState(godo.ActionCompleted, imageId, action.ID, client, timeout); err != nil {
			if err != nil {
				return fmt.Errorf("Error transferring image: %s", err)
			}
		}
	}

	return nil
}

func deleteImageFromSpaces(p *PostProcessor, s *session.Session) (err error) {
	s3conn := s3.New(s)
	_, err = s3conn.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &p.config.SpaceName,
		Key:    &p.config.ObjectName,
	})
	if err != nil {
		return fmt.Errorf("Failed to delete spaces://%s/%s: %s", p.config.SpaceName, p.config.ObjectName, err)
	}

	return nil
}
