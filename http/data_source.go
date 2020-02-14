package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func httpDataSource() *schema.Resource {

	var allowedMethods = []string{"GET", "POST", "PATCH", "DELETE", "PUT", "HEAD", "OPTIONS", "CONNECT", "TRACE"}

	return &schema.Resource{
		Read: dataSourceRead,

		Schema: map[string]*schema.Schema{
			"method": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "GET",
				ValidateFunc: validation.StringInSlice(allowedMethods, false),
			},

			"url": {
				Type:     schema.TypeString,
				Required: true,
			},

			"request_headers": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Sensitive: true,
			},

			"request_body": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},

			"response_status_code": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      200,
				ValidateFunc: validation.IntBetween(100, 599),
			},

			"body": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"headers": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceRead(d *schema.ResourceData, meta interface{}) error {

	url := d.Get("url").(string)
	method := d.Get("method").(string)
	headers := d.Get("request_headers").(map[string]interface{})
	body := d.Get("request_body").(string)
	statusCode := d.Get("response_status_code").(int)

	client := &http.Client{}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("Error creating request: %s", err)
	}

	for name, value := range headers {
		req.Header.Set(name, value.(string))
	}

	if len(body) != 0 {
		req.Body = ioutil.NopCloser(strings.NewReader(body))
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error making a request: %s", err)
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error while reading response body. %s", err)
	}

	if resp.StatusCode != statusCode {
		return fmt.Errorf("HTTP request error. Response code: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" || isContentTypeAllowed(contentType) == false {
		return fmt.Errorf("Content-Type is not a text type. Got: %s", contentType)
	}

	d.Set("body", string(bytes))
	d.Set("headers", flattenResponseHeaders(resp.Header))
	d.SetId(uuid.New().String())

	return nil
}
