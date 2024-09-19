//go:generate packer-sdc mapstructure-to-hcl2 -type Config,DatasourceOutput
//go:generate packer-sdc struct-markdown
package image

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"

	builder "github.com/digitalocean/packer-plugin-digitalocean/builder/digitalocean"
	"github.com/digitalocean/packer-plugin-digitalocean/version"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/useragent"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/oauth2"
)

var (
	validImageTypes = []string{"application", "distribution", "user"}
)

type Config struct {
	// The API token to used to access your account. It can also be specified via
	// the DIGITALOCEAN_TOKEN or DIGITALOCEAN_ACCESS_TOKEN environment variables.
	APIToken string `mapstructure:"api_token" required:"true"`
	// A non-standard API endpoint URL. Set this if you are  using a DigitalOcean API
	// compatible service. It can also be specified via environment variable DIGITALOCEAN_API_URL.
	APIURL string `mapstructure:"api_url"`
	// The maximum number of retries for requests that fail with a 429 or 500-level error.
	// The default value is 5. Set to 0 to disable reties.
	HTTPRetryMax *int `mapstructure:"http_retry_max" required:"false"`
	// The maximum wait time (in seconds) between failed API requests. Default: 30.0
	HTTPRetryWaitMax *float64 `mapstructure:"http_retry_wait_max" required:"false"`
	// The minimum wait time (in seconds) between failed API requests. Default: 1.0
	HTTPRetryWaitMin *float64 `mapstructure:"http_retry_wait_min" required:"false"`
	// The name of the image to return. Only one of `name` or `name_regex` may be provided.
	Name string `mapstructure:"name"`
	// A regex matching the name of the image to return. Only one of `name` or `name_regex` may be provided.
	NameRegex string `mapstructure:"name_regex"`
	// Filter the images searched by type. This may be one of `application`, `distribution`, or `user`.
	// By default, all image types are searched.
	Type string `mapstructure:"type"`
	// A DigitalOcean region slug (e.g. `nyc3`). When provided, only images available in that region
	// will be returned.
	Region string `mapstructure:"region"`
	// A boolean value determining how to handle multiple matching images. By default, multiple matching images
	// results in an error. When set to `true`, the most recently created image is returned instead.
	Latest bool `mapstructure:"latest"`
}

type Datasource struct {
	config Config
}

type DatasourceOutput struct {
	// The ID of the found image.
	ImageID int `mapstructure:"image_id"`
	// The regions the found image is availble in.
	ImageRegions []string `mapstructure:"image_regions"`
}

func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Configure(raws ...interface{}) error {
	err := config.Decode(&d.config, nil, raws...)
	if err != nil {
		return err
	}

	var errs *packersdk.MultiError

	if d.config.APIToken == "" {
		d.config.APIToken = os.Getenv("DIGITALOCEAN_TOKEN")
		if d.config.APIToken == "" {
			d.config.APIToken = os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
		}
	}
	if d.config.APIURL == "" {
		d.config.APIURL = os.Getenv("DIGITALOCEAN_API_URL")
	}

	if d.config.HTTPRetryMax == nil {
		d.config.HTTPRetryMax = godo.PtrTo(5)
		if max := os.Getenv("DIGITALOCEAN_HTTP_RETRY_MAX"); max != "" {
			maxInt, err := strconv.Atoi(max)
			if err != nil {
				return err
			}
			d.config.HTTPRetryMax = godo.PtrTo(maxInt)
		}
	}
	if d.config.HTTPRetryWaitMax == nil {
		d.config.HTTPRetryWaitMax = godo.PtrTo(30.0)
		if waitMax := os.Getenv("DIGITALOCEAN_HTTP_RETRY_WAIT_MAX"); waitMax != "" {
			waitMaxFloat, err := strconv.ParseFloat(waitMax, 64)
			if err != nil {
				return err
			}
			d.config.HTTPRetryWaitMax = godo.PtrTo(waitMaxFloat)
		}
	}
	if d.config.HTTPRetryWaitMin == nil {
		d.config.HTTPRetryWaitMin = godo.PtrTo(1.0)
		if waitMin := os.Getenv("DIGITALOCEAN_HTTP_RETRY_WAIT_MIN"); waitMin != "" {
			waitMinFloat, err := strconv.ParseFloat(waitMin, 64)
			if err != nil {
				return err
			}
			d.config.HTTPRetryWaitMin = godo.PtrTo(waitMinFloat)
		}
	}

	if d.config.APIToken == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("api_token is required"))
	}

	if d.config.Name == "" && d.config.NameRegex == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("one of name or name_regex is required"))
	}

	if d.config.Name != "" && d.config.NameRegex != "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("only one of name or name_regex can be set"))
	}

	if d.config.Type != "" {
		if !contains(validImageTypes, d.config.Type) {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("invalid type; must be one of: %v", validImageTypes))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	ua := useragent.String(version.PluginVersion.FormattedVersion())
	clientOpts := []godo.ClientOpt{godo.SetUserAgent(ua)}
	if d.config.APIURL != "" {
		_, err := url.Parse(d.config.APIURL)
		if err != nil {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("invalid API URL, %s.", err)
		}

		clientOpts = append(clientOpts, godo.SetBaseURL(d.config.APIURL))
	}

	if *d.config.HTTPRetryMax > 0 {
		clientOpts = append(clientOpts, godo.WithRetryAndBackoffs(godo.RetryConfig{
			RetryMax:     *d.config.HTTPRetryMax,
			RetryWaitMin: d.config.HTTPRetryWaitMin,
			RetryWaitMax: d.config.HTTPRetryWaitMax,
			Logger:       log.Default(),
		}))
	}

	oauthClient := oauth2.NewClient(context.TODO(), &builder.APITokenSource{
		AccessToken: d.config.APIToken,
	})
	client, err := godo.New(oauthClient, clientOpts...)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}

	opts := &godo.ListOptions{
		Page:    1,
		PerPage: 200,
	}

	imageListFunc := client.Images.List
	switch d.config.Type {
	case "user":
		imageListFunc = client.Images.ListUser
	case "application":
		imageListFunc = client.Images.ListApplication
	case "distribution":
		imageListFunc = client.Images.ListDistribution
	}

	var imageList []godo.Image
	for {
		images, resp, err := imageListFunc(context.Background(), opts)

		if err != nil {
			return cty.NullVal(cty.EmptyObject), err
		}

		imageList = append(imageList, images...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return cty.NullVal(cty.EmptyObject), err
		}

		opts.Page = page + 1
	}

	result, err := filterImages(&d.config, imageList)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), err
	}

	output := DatasourceOutput{
		ImageID:      result.ID,
		ImageRegions: result.Regions,
	}

	log.Printf("[DEBUG] found image: %v", result.ID)

	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}

func filterImages(c *Config, images []godo.Image) (godo.Image, error) {
	result := make([]godo.Image, 0)
	if c.Name != "" {
		result = filterByName(images, c.Name)
	}

	if c.NameRegex != "" {
		result = filterByNameRegex(images, c.NameRegex)
	}

	if c.Region != "" {
		result = filterByRegion(result, c.Region)
	}

	if len(result) > 1 {
		if c.Latest {
			return findLatest(result), nil
		}

		return godo.Image{}, fmt.Errorf("More than one matching image found: %v", result)
	}
	if len(result) == 0 {
		return godo.Image{}, errors.New("No matching image found")
	}

	return result[0], nil
}

func filterByName(images []godo.Image, name string) []godo.Image {
	result := make([]godo.Image, 0)
	for _, i := range images {
		if i.Name == name {
			result = append(result, i)
		}
	}

	return result
}

func filterByNameRegex(images []godo.Image, name string) []godo.Image {
	r := regexp.MustCompile(name)
	result := make([]godo.Image, 0)
	for _, i := range images {
		if r.MatchString(i.Name) {
			result = append(result, i)
		}
	}

	return result
}

func filterByRegion(images []godo.Image, region string) []godo.Image {
	result := make([]godo.Image, 0)
	for _, i := range images {
		for _, r := range i.Regions {
			if r == region {
				result = append(result, i)
				break
			}
		}
	}

	return result
}

func findLatest(images []godo.Image) godo.Image {
	sort.Slice(images, func(i, j int) bool {
		itime, _ := time.Parse(time.RFC3339, images[i].Created)
		jtime, _ := time.Parse(time.RFC3339, images[j].Created)
		return itime.Unix() > jtime.Unix()
	})

	return images[0]
}

func contains(list []string, term string) bool {
	for _, t := range list {
		if t == term {
			return true
		}
	}
	return false
}
