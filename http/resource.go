package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func httpResource() *schema.Resource {

	var allowedMethods = []string{"GET", "POST", "PATCH", "DELETE", "PUT", "HEAD", "OPTIONS", "CONNECT", "TRACE"}

	createSchema := map[string]*schema.Schema{
		"url": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},

		"method": {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			Default:      "POST",
			ValidateFunc: validation.StringInSlice(allowedMethods, false),
		},

		"response_status_code": {
			Type:         schema.TypeInt,
			Optional:     true,
			ForceNew:     true,
			Default:      200,
			ValidateFunc: validation.IntBetween(100, 599),
		},

		"request_headers": {
			Type:      schema.TypeMap,
			Optional:  true,
			ForceNew:  true,
			Sensitive: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},

		"request_body": {
			Type:      schema.TypeString,
			Optional:  true,
			ForceNew:  true,
			Sensitive: true,
		},

		"body": {
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: true,
		},

		"headers": {
			Type:      schema.TypeMap,
			Computed:  true,
			Sensitive: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}

	updateSchema := map[string]*schema.Schema{
		"url": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},

		"method": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice(allowedMethods, false),
		},

		"response_status_code": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      200,
			ValidateFunc: validation.IntBetween(100, 599),
		},

		"request_headers": {
			Type:      schema.TypeMap,
			Optional:  true,
			Sensitive: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},

		"request_body": {
			Type:      schema.TypeString,
			Optional:  true,
			Sensitive: true,
		},

		"body": {
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: true,
		},

		"headers": {
			Type:      schema.TypeMap,
			Computed:  true,
			Sensitive: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}

	deleteSchema := map[string]*schema.Schema{
		"url": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},

		"method": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice(allowedMethods, false),
		},

		"response_status_code": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      200,
			ValidateFunc: validation.IntBetween(100, 599),
		},

		"request_headers": {
			Type:      schema.TypeMap,
			Optional:  true,
			Sensitive: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},

		"request_body": {
			Type:      schema.TypeString,
			Optional:  true,
			Sensitive: true,
		},
	}

	return &schema.Resource{
		Create: resourceCreate,
		Read:   func(*schema.ResourceData, interface{}) error { return nil },
		Update: resourceUpdate,
		Delete: resourceDelete,

		Schema: map[string]*schema.Schema{
			"triggers": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: true,
			},

			"action": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"create": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: createSchema,
							},
						},
						"update": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: updateSchema,
							},
						},
						"delete": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: deleteSchema,
							},
						},
					},
				},
			},
		},
	}
}

func resourceCreate(d *schema.ResourceData, meta interface{}) error {
	return httpRequest(d, meta, "create")
}

func resourceUpdate(d *schema.ResourceData, meta interface{}) error {
	return httpRequest(d, meta, "update")
}

func resourceDelete(d *schema.ResourceData, meta interface{}) error {
	return httpRequest(d, meta, "delete")
}

func httpRequest(d *schema.ResourceData, meta interface{}, action string) error {
	if action == "update" && !d.HasChange("action.0."+action+".0") {
		// nothing to do
		return nil
	}

	method := d.Get("action.0." + action + ".0.method").(string)
	if len(method) == 0 {
		d.Set("action", flattenAction(d.Get("action"), []byte{}, http.Header{}, action))
		if action == "create" {
			d.SetId(time.Now().UTC().String())
		}
		return nil
	}

	url := d.Get("action.0." + action + ".0.url").(string)

	headers := d.Get("action.0." + action + ".0.request_headers").(map[string]interface{})
	body := d.Get("action.0." + action + ".0.request_body").(string)
	statusCode := d.Get("action.0." + action + ".0.response_status_code").(int)

	client := &http.Client{}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("Error creating %s request: %s", action, err)
	}

	for name, value := range headers {
		req.Header.Set(name, value.(string))
	}

	if len(body) != 0 {
		req.Body = ioutil.NopCloser(strings.NewReader(body))
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error during making a %s request: %s", action, url)
	}

	defer resp.Body.Close()

	if action != "delete" {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Error while reading %s response body. %s", action, err)
		}

		if resp.StatusCode != statusCode {
			return fmt.Errorf("%s HTTP request error. Response code: %d\n\n%s", action, resp.StatusCode, string(bytes))
		}

		// ignore responses from the delete action
		contentType := resp.Header.Get("Content-Type")
		if contentType == "" || isContentTypeAllowed(contentType) == false {
			return fmt.Errorf("Content-Type is not a text type. Got: %s", contentType)
		}

		d.Set("action", flattenAction(d.Get("action"), bytes, resp.Header, action))
	} else {
		if resp.StatusCode != statusCode {
			return fmt.Errorf("%s HTTP request error. Response code: %d", action, resp.StatusCode)
		}
	}

	if action == "create" {
		d.SetId(uuid.New().String())
	}

	return nil
}
